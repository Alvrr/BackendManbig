package controllers

import (
	"backend/models"
	"backend/repository"
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
	if role != "admin" && t.KasirID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Akses ditolak"})
	}
	return c.JSON(t)
}

// POST /transaksi (admin+kasir)
func CreateTransaksi(c *fiber.Ctx) error {
	var t models.Transaksi
	if err := c.BodyParser(&t); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Data tidak valid"})
	}
	if t.KasirID == "" || t.PelangganID == "" || t.TotalHarga <= 0 {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"message": "kasir_id, pelanggan_id, total_harga wajib"})
	}

	// Enrich item names if missing
	if len(t.Items) > 0 {
		for i := range t.Items {
			if t.Items[i].NamaProduk == "" && t.Items[i].ProdukID != "" {
				if p, err := repository.GetProdukByID(t.Items[i].ProdukID); err == nil && p != nil {
					t.Items[i].NamaProduk = p.NamaProduk
				}
			}
		}
	}

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
	t, err := repository.GetTransaksiByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Transaksi tidak ditemukan"})
	}
	if role != "admin" && t.KasirID != userID {
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
	t, err := repository.GetTransaksiByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Transaksi tidak ditemukan"})
	}
	if role != "admin" && t.KasirID != userID {
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
