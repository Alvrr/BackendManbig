package controllers

import (
	"backend/models"
	"backend/repository"
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
	return c.JSON(fiber.Map{"message": "Transaksi berhasil dihapus"})
}
