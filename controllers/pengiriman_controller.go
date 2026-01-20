package controllers

import (
	"backend/models"
	"backend/repository"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

// List pengiriman: admin/kasir view all, driver only own
func GetAllPengiriman(c *fiber.Ctx) error {
	role := c.Locals("userRole").(string)
	id := c.Locals("userID").(string)
	filter := bson.M{}
	if role == "driver" {
		// IMPORTANT: driver hanya boleh melihat pengiriman miliknya sendiri
		// (field DB: driver_id  diperlakukan sebagai assigned_driver_id)
		filter["driver_id"] = id
	}
	if role == "kasir" {
		// IMPORTANT: kasir hanya boleh melihat pengiriman untuk transaksi miliknya sendiri
		// agar dashboard/pengiriman tidak bocor antar kasir.
		trx, err := repository.ListTransaksi(bson.M{"kasir_id": id}, 0, 0)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"message": "Gagal ambil transaksi kasir", "error": err.Error()})
		}
		ids := make([]string, 0, len(trx))
		for _, t := range trx {
			if t.ID != "" {
				ids = append(ids, t.ID)
			}
		}
		if len(ids) == 0 {
			return c.JSON([]models.Pengiriman{})
		}
		filter["transaksi_id"] = bson.M{"$in": ids}
	}
	list, err := repository.GetPengirimanFiltered(filter)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal ambil data", "error": err.Error()})
	}
	return c.JSON(list)
}

func GetPengirimanByID(c *fiber.Ctx) error {
	id := c.Params("id")
	role := c.Locals("userRole").(string)
	userID := c.Locals("userID").(string)
	data, err := repository.GetPengirimanByID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "Data tidak ditemukan"})
	}
	if role == "driver" && data.DriverID != userID {
		// IMPORTANT: driver tidak boleh akses pengiriman driver lain
		return c.Status(403).JSON(fiber.Map{"message": "Akses ditolak"})
	}
	if role == "kasir" {
		// IMPORTANT: kasir tidak boleh akses pengiriman milik kasir lain
		trx, err := repository.GetTransaksiByID(data.TransaksiID)
		if err != nil || trx == nil || trx.KasirID != userID {
			return c.Status(403).JSON(fiber.Map{"message": "Akses ditolak"})
		}
	}
	// Enrich detail: ambil transaksi dan pelanggan terkait untuk keperluan tampilan detail
	var trx *models.Transaksi
	if data.TransaksiID != "" {
		if t, err := repository.GetTransaksiByID(data.TransaksiID); err == nil {
			trx = t
		}
	}
	var pelangganNama string
	var pelangganID string
	var items []models.TransaksiItem
	var totalToko float64
	if trx != nil {
		pelangganID = trx.PelangganID
		if p, err := repository.GetPelangganByID(trx.PelangganID); err == nil && p != nil {
			pelangganNama = p.Nama
		}
		items = trx.Items
		if trx.TotalHarga > 0 {
			totalToko = trx.TotalHarga
		} else {
			// hitung dari items jika total_harga belum terisi
			var sum float64
			for _, it := range trx.Items {
				sum += float64(it.Jumlah) * it.Harga
			}
			totalToko = sum
		}
	}
	return c.JSON(fiber.Map{
		"id":           data.ID,
		"transaksi_id": data.TransaksiID,
		"driver_id":    data.DriverID,
		"jenis":        data.Jenis,
		"ongkir":       data.Ongkir,
		"status":       data.Status,
		"alasan_batal": data.AlasanBatal,
		"created_at":   data.CreatedAt,
		// Enriched fields for detail dialog
		"pelanggan_id":   pelangganID,
		"pelanggan_nama": pelangganNama,
		"total_toko":     totalToko, // total belanja tanpa ongkir
		"items":          items,
	})
}

// Create: admin/kasir set driver assignment
func CreatePengiriman(c *fiber.Ctx) error {
	var p models.Pengiriman
	if err := c.BodyParser(&p); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Request tidak valid", "error": err.Error()})
	}
	if p.TransaksiID == "" || p.DriverID == "" {
		return c.Status(422).JSON(fiber.Map{"message": "transaksi_id dan driver_id wajib"})
	}
	// Hitung ongkir server-side berdasarkan jenis kendaraan
	jenis := strings.ToLower(strings.TrimSpace(p.Jenis))
	if jenis == "" {
		jenis = "mobil"
	}
	if jenis == "mobil" {
		p.Ongkir = 25000
	} else if jenis == "motor" {
		p.Ongkir = 10000
	} else {
		return c.Status(422).JSON(fiber.Map{"message": "jenis harus salah satu dari: mobil, motor"})
	}
	p.Jenis = jenis
	id, err := repository.GenerateID("pengiriman")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal generate ID", "error": err.Error()})
	}
	p.ID = id
	if p.Status == "" {
		p.Status = "diproses"
	}
	p.CreatedAt = time.Now()
	res, err := repository.CreatePengiriman(p)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal simpan", "error": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"message": "Berhasil ditambahkan", "data": res.InsertedID})
}

// Update: driver can update own status; admin/kasir can update any
func UpdatePengiriman(c *fiber.Ctx) error {
	id := c.Params("id")
	role := c.Locals("userRole").(string)
	userID := c.Locals("userID").(string)
	existing, err := repository.GetPengirimanByID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "Data tidak ditemukan"})
	}
	if role == "driver" && existing.DriverID != userID {
		// IMPORTANT: driver hanya boleh update status pengiriman miliknya sendiri
		return c.Status(403).JSON(fiber.Map{"message": "Akses ditolak"})
	}
	if role == "kasir" {
		// IMPORTANT: kasir hanya boleh update pengiriman untuk transaksi miliknya sendiri
		trx, err := repository.GetTransaksiByID(existing.TransaksiID)
		if err != nil || trx == nil || trx.KasirID != userID {
			return c.Status(403).JSON(fiber.Map{"message": "Akses ditolak"})
		}
	}
	var payload models.Pengiriman
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Request tidak valid", "error": err.Error()})
	}
	upd, err := repository.UpdatePengiriman(id, payload)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal update", "error": err.Error()})
	}

	// Sinkronkan status ke transaksi & pembayaran
	statusLower := strings.ToLower(payload.Status)
	if statusLower == "selesai" || statusLower == "dikirim" || statusLower == "sedang diantarkan" || statusLower == "sedang di antar" {
		// Map ke status transaksi yang ramah tampil
		tStatus := ""
		if statusLower == "selesai" {
			tStatus = "Selesai"
		} else {
			// untuk "dikirim" maupun "sedang diantarkan" gunakan label berikut
			tStatus = "Sedang Diantarkan"
		}
		// Update status transaksi
		_, _ = repository.UpdateTransaksi(existing.TransaksiID, bson.M{"status": tStatus})

		// Jika selesai: tandai pembayaran menjadi Selesai
		if tStatus == "Selesai" {
			pays, _ := repository.GetPembayaranFiltered(bson.M{"transaksi_id": existing.TransaksiID})
			for _, p := range pays {
				_, _ = repository.UpdatePembayaran(p.ID, models.Pembayaran{Status: "Selesai"})
			}
			// Perbarui keterangan mutasi stok dari 'reservasi' menjadi 'terjual'
			_ = repository.UpdateMutasiKeteranganByRef(existing.TransaksiID, "terjual")
		}
	}

	// Jika batal: wajib simpan alasan, transaksi kembali ke proses, pembayaran ditandai Batal
	if statusLower == "batal" {
		// Kembalikan transaksi ke status 'Proses'
		_, _ = repository.UpdateTransaksi(existing.TransaksiID, bson.M{"status": "Proses"})
		// Tandai semua pembayaran terkait sebagai 'Batal'
		pays, _ := repository.GetPembayaranFiltered(bson.M{"transaksi_id": existing.TransaksiID})
		for _, p := range pays {
			_, _ = repository.UpdatePembayaran(p.ID, models.Pembayaran{Status: "Batal"})
		}
	}
	return c.JSON(fiber.Map{"message": "Berhasil diupdate", "modified": upd.ModifiedCount})
}

// Delete: admin/kasir only
func DeletePengiriman(c *fiber.Ctx) error {
	id := c.Params("id")
	res, err := repository.DeletePengiriman(id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal hapus", "error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Berhasil dihapus", "deleted": res.DeletedCount})
}
