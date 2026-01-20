package controllers

import (
	"backend/models"
	"backend/repository"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

// GET /stok/saldo/:produk_id (view semua role)
func GetSaldoProduk(c *fiber.Ctx) error {
	id := c.Params("produk_id")
	s, err := repository.GetSaldoProduk(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal menghitung saldo"})
	}
	return c.JSON(s)
}

// GET /stok/mutasi/:produk_id (view semua role)
func GetMutasiByProduk(c *fiber.Ctx) error {
	id := c.Params("produk_id")
	list, err := repository.GetMutasiByProduk(id, 0, 0)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal mengambil mutasi"})
	}
	return c.JSON(list)
}

// GET /stok/mutasi (global list with filters) - view semua role
func ListMutasi(c *fiber.Ctx) error {
	// Query params: produk_id, jenis, keterangan, ref_type, ref_id, start, end, page, page_size
	filter := bson.M{}
	produkID := c.Query("produk_id")
	jenis := c.Query("jenis")
	ket := c.Query("keterangan")
	refType := c.Query("ref_type")
	refID := c.Query("ref_id")
	if produkID != "" {
		filter["produk_id"] = produkID
	}
	if jenis != "" {
		filter["jenis"] = jenis
	}
	if ket != "" {
		filter["keterangan"] = ket
	}
	if refType != "" {
		filter["ref_type"] = refType
	}
	if refID != "" {
		filter["ref_id"] = refID
	}
	// Date range
	start := c.Query("start")
	end := c.Query("end")
	if start != "" || end != "" {
		// created_at between start and end
		rangeFilter := bson.M{}
		if start != "" {
			if t, err := time.Parse(time.RFC3339, start); err == nil {
				rangeFilter["$gte"] = t
			}
		}
		if end != "" {
			if t, err := time.Parse(time.RFC3339, end); err == nil {
				rangeFilter["$lte"] = t
			}
		}
		if len(rangeFilter) > 0 {
			filter["created_at"] = rangeFilter
		}
	}
	page, _ := strconv.Atoi(c.Query("page", "0"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "0"))
	sortDesc := c.Query("sort", "desc") != "asc"
	list, err := repository.ListMutasi(filter, page, pageSize, sortDesc)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal mengambil mutasi"})
	}
	return c.JSON(list)
}

// POST /stok (admin+gudang)
func CreateMutasi(c *fiber.Ctx) error {
	var m models.StokMutasi
	if err := c.BodyParser(&m); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Data tidak valid"})
	}
	if m.ProdukID == "" || m.Jenis == "" {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"message": "produk_id dan jenis wajib"})
	}
	// Validasi jenis mutasi (hanya masuk/keluar)
	if m.Jenis != "masuk" && m.Jenis != "keluar" {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"message": "jenis harus salah satu dari: masuk, keluar"})
	}
	// Aturan jumlah: masuk/keluar wajib > 0
	if m.Jumlah <= 0 {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"message": "jumlah harus lebih dari 0 untuk mutasi masuk/keluar"})
	}

	// Jika keluar (manual), validasi stok mencukupi lebih dulu
	if m.Jenis == "keluar" {
		saldo, err := repository.GetSaldoProduk(m.ProdukID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal membaca saldo"})
		}
		if m.Jumlah > saldo.Saldo {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Stok tidak mencukupi"})
		}
	}
	// Normalisasi untuk adjust: ubah menjadi delta masuk/keluar dibanding saldo saat ini
	// Jenis adjust dihapus: tidak ada normalisasi target saldo
	// Set ID dari counter stok + created_at
	id, err := repository.GenerateID("stok")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal generate ID"})
	}
	m.ID = id
	// Wajib isi user_id dari token, jika tidak ada tolak
	uid, ok := c.Locals("userID").(string)
	if !ok || uid == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "User login tidak valid"})
	}
	m.UserID = uid
	// Mutasi manual: jika ref_type kosong, set default 'manual'.
	if m.RefType == "" {
		m.RefType = "manual"
	}
	// Auto set ref_id untuk mutasi manual jika belum diisi: gunakan ID mutasi
	if m.RefType == "manual" && m.RefID == "" {
		m.RefID = m.ID
	}
	m.CreatedAt = time.Now()
	if _, err := repository.CreateMutasi(&m); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Mutasi stok berhasil dibuat", "id": m.ID})
}
