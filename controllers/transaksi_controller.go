package controllers

import (
	"backend/models"
	"backend/repository"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

// GET /transaksi (admin semua; kasir hanya miliknya)
func ListTransaksi(c *fiber.Ctx) error {
	role, _ := c.Locals("userRole").(string)
	userID, _ := c.Locals("userID").(string)
	filter := bson.M{}
	if role != "admin" {
		// IMPORTANT: kasir hanya boleh melihat transaksi miliknya sendiri
		// (field DB: kasir_id  diperlakukan sebagai created_by)
		filter["kasir_id"] = userID
	}
	list, err := repository.ListTransaksi(filter, 0, 0)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal mengambil transaksi"})
	}
	return c.JSON(list)
}

// GET /transaksi/:id (admin; kasir hanya jika miliknya)
func GetTransaksiByID(c *fiber.Ctx) error {
	id := c.Params("id")
	role, _ := c.Locals("userRole").(string)
	userID, _ := c.Locals("userID").(string)
	t, err := repository.GetTransaksiByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Transaksi tidak ditemukan"})
	}
	// IMPORTANT: kasir hanya boleh melihat transaksi miliknya sendiri
	if role == "kasir" && t.KasirID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Akses ditolak"})
	}
	// FINAL RULE: selain admin/kasir tidak boleh akses transaksi
	if role != "admin" && role != "kasir" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Akses ditolak"})
	}
	return c.JSON(t)
}

// POST /transaksi (admin+kasir)
func CreateTransaksi(c *fiber.Ctx) error {
	role, _ := c.Locals("userRole").(string)
	userID, _ := c.Locals("userID").(string)
	if role != "kasir" {
		// Admin/gudang/driver ditolak untuk create transaksi
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Akses ditolak"})
	}

	var body struct {
		PelangganID string `json:"pelanggan_id"`
		Items       []struct {
			ProdukID string `json:"produk_id"`
			Jumlah   int    `json:"jumlah"`
		} `json:"items"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Data tidak valid"})
	}
	if body.PelangganID == "" {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"message": "pelanggan_id wajib"})
	}
	if len(body.Items) == 0 {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"message": "items wajib"})
	}

	// Aggregate qty per produk untuk mencegah bypass stok via split items
	aggQty := map[string]int{}
	for _, it := range body.Items {
		if it.ProdukID == "" {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"message": "produk_id wajib"})
		}
		if it.Jumlah <= 0 {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"message": "jumlah harus lebih dari 0"})
		}
		aggQty[it.ProdukID] += it.Jumlah
	}

	// Validasi stok di backend (qty <= stok) dan ambil harga produk dari DB
	produkCache := map[string]*models.Produk{}
	for produkID, qty := range aggQty {
		saldo, err := repository.GetSaldoProduk(produkID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal membaca saldo"})
		}
		if qty > saldo.Saldo {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": fmt.Sprintf("Stok tidak mencukupi untuk produk %s", produkID)})
		}
		p, err := repository.GetProdukByID(produkID)
		if err != nil || p == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": fmt.Sprintf("Produk tidak ditemukan: %s", produkID)})
		}
		produkCache[produkID] = p
	}

	// Build transaksi: kasir_id dari JWT, harga dari DB, totals dihitung server-side
	t := models.Transaksi{
		KasirID:     userID,
		PelangganID: body.PelangganID,
		Status:      "proses",
		Items:       []models.TransaksiItem{},
	}

	var totalProduk int
	var totalHarga float64
	for _, it := range body.Items {
		p := produkCache[it.ProdukID]
		harga := float64(0)
		nama := ""
		if p != nil {
			harga = p.HargaJual
			nama = p.NamaProduk
		}
		qty := it.Jumlah
		item := models.TransaksiItem{
			ProdukID:   it.ProdukID,
			NamaProduk: nama,
			Jumlah:     qty,
			Harga:      harga,
		}
		t.Items = append(t.Items, item)
		totalProduk += qty
		totalHarga += harga * float64(qty)
	}
	t.TotalProduk = totalProduk
	t.TotalHarga = totalHarga

	id, err := repository.GenerateID("transaksi")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal generate ID"})
	}
	t.ID = id
	t.CreatedAt = time.Now()
	if _, err := repository.CreateTransaksi(&t); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal membuat transaksi"})
	}
	// Kurangi stok (reservasi): buat mutasi keluar untuk setiap item transaksi
	// Mutasi dibuat dengan user_id dari kasir pembuat transaksi dan ditandai ref transaksi
	if len(t.Items) > 0 {
		for _, it := range t.Items {
			if it.ProdukID == "" || it.Jumlah <= 0 {
				continue
			}
			// Generate ID unik untuk setiap mutasi agar tidak terjadi duplicate key (_id)
			mutasiID, genErr := repository.GenerateID("stok")
			if genErr != nil {
				// Jika gagal generate ID, tetap coba insert tanpa _id (Mongo akan menolak duplikat jika kosong)
				// Namun kita tetap lanjut agar transaksi tidak gagal seluruhnya
			}
			m := &models.StokMutasi{
				ID:         mutasiID,
				ProdukID:   it.ProdukID,
				Jenis:      "keluar",
				Jumlah:     it.Jumlah,
				UserID:     t.KasirID,
				RefID:      t.ID,
				RefType:    "transaksi",
				Keterangan: "reservasi",
				CreatedAt:  time.Now(),
			}
			// Abaikan error agar transaksi tetap sukses. Log di repository jika perlu.
			_, _ = repository.CreateMutasi(m)
		}
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Transaksi berhasil dibuat", "id": t.ID})
}

// PUT /transaksi/:id (admin+kasir; kasir hanya miliknya)
func UpdateTransaksi(c *fiber.Ctx) error {
	id := c.Params("id")
	role, _ := c.Locals("userRole").(string)
	userID, _ := c.Locals("userID").(string)
	if role != "kasir" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Akses ditolak"})
	}
	t, err := repository.GetTransaksiByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Transaksi tidak ditemukan"})
	}
	if t.KasirID != userID {
		// IMPORTANT: kasir tidak boleh update transaksi kasir lain
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Akses ditolak"})
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Data tidak valid"})
	}
	if body.Status == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Status wajib"})
	}
	if _, err := repository.UpdateTransaksi(id, bson.M{"status": body.Status}); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal update transaksi"})
	}
	// Jika dibatalkan, kembalikan stok dengan mutasi masuk
	if strings.EqualFold(body.Status, "batal") && len(t.Items) > 0 {
		for _, it := range t.Items {
			if it.ProdukID == "" || it.Jumlah <= 0 {
				continue
			}
			mutasiID, _ := repository.GenerateID("stok")
			m := &models.StokMutasi{
				ID:         mutasiID,
				ProdukID:   it.ProdukID,
				Jenis:      "masuk",
				Jumlah:     it.Jumlah,
				UserID:     t.KasirID,
				RefID:      t.ID,
				RefType:    "transaksi",
				Keterangan: "batal",
				CreatedAt:  time.Now(),
			}
			_, _ = repository.CreateMutasi(m)
		}
	}
	return c.JSON(fiber.Map{"message": "Transaksi berhasil diupdate"})
}

// DELETE /transaksi/:id (admin+kasir; kasir hanya miliknya)
func DeleteTransaksi(c *fiber.Ctx) error {
	id := c.Params("id")
	role, _ := c.Locals("userRole").(string)
	userID, _ := c.Locals("userID").(string)
	if role != "kasir" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Akses ditolak"})
	}
	t, err := repository.GetTransaksiByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Transaksi tidak ditemukan"})
	}
	if t.KasirID != userID {
		// IMPORTANT: kasir tidak boleh hapus transaksi kasir lain
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Akses ditolak"})
	}
	if _, err := repository.DeleteTransaksi(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal hapus transaksi"})
	}
	// Kembalikan stok jika transaksi memiliki item (asumsi reservasi sudah keluar)
	if len(t.Items) > 0 {
		for _, it := range t.Items {
			if it.ProdukID == "" || it.Jumlah <= 0 {
				continue
			}
			mutasiID, _ := repository.GenerateID("stok")
			m := &models.StokMutasi{
				ID:         mutasiID,
				ProdukID:   it.ProdukID,
				Jenis:      "masuk",
				Jumlah:     it.Jumlah,
				UserID:     t.KasirID,
				RefID:      t.ID,
				RefType:    "transaksi",
				Keterangan: "hapus",
				CreatedAt:  time.Now(),
			}
			_, _ = repository.CreateMutasi(m)
		}
	}
	return c.JSON(fiber.Map{"message": "Transaksi berhasil dihapus"})
}
