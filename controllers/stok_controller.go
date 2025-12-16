package controllers

import (
	"backend/models"
	"backend/repository"
	"time"

	"github.com/gofiber/fiber/v2"
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

// POST /stok (admin+gudang)
func CreateMutasi(c *fiber.Ctx) error {
	var m models.StokMutasi
	if err := c.BodyParser(&m); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Data tidak valid"})
	}
	if m.ProdukID == "" || m.Jenis == "" || m.Jumlah == 0 {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"message": "produk_id, jenis, jumlah wajib"})
	}
	// Normalisasi untuk adjust: ubah menjadi delta masuk/keluar dibanding saldo saat ini
	if m.Jenis == "adjust" {
		saldo, err := repository.GetSaldoProduk(m.ProdukID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal membaca saldo untuk adjust"})
		}
		// Jika sama, tidak perlu perubahan; jadikan jumlah 0 masuk agar tidak mempengaruhi saldo
		if m.Jumlah == saldo.Saldo {
			m.Jenis = "masuk"
			m.Jumlah = 0
		} else if m.Jumlah > saldo.Saldo {
			m.Jenis = "masuk"
			m.Jumlah = m.Jumlah - saldo.Saldo
		} else {
			m.Jenis = "keluar"
			m.Jumlah = saldo.Saldo - m.Jumlah
		}
	}
	// Set ID dari counter stok + created_at
	id, err := repository.GenerateID("stok")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal generate ID"})
	}
	m.ID = id
	m.CreatedAt = time.Now()
	if _, err := repository.CreateMutasi(&m); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Mutasi stok berhasil dibuat", "id": m.ID})
}
