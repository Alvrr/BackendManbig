package controllers

import (
	"backend/models"
	"backend/repository"
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
		filter["driver_id"] = id
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
		return c.Status(403).JSON(fiber.Map{"message": "Akses ditolak"})
	}
	return c.JSON(data)
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
		return c.Status(403).JSON(fiber.Map{"message": "Akses ditolak"})
	}
	var payload models.Pengiriman
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Request tidak valid", "error": err.Error()})
	}
	upd, err := repository.UpdatePengiriman(id, payload)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Gagal update", "error": err.Error()})
	}

	// Sinkronkan status ke transaksi & pembayaran jika selesai/dikirim
	if payload.Status == "selesai" || payload.Status == "dikirim" {
		tStatus := payload.Status
		if payload.Status == "selesai" {
			tStatus = "Selesai"
		}
		if payload.Status == "dikirim" {
			tStatus = "Dikirim"
		}
		// Update field status pada transaksi terkait
		_, _ = repository.UpdateTransaksi(existing.TransaksiID, bson.M{"status": tStatus})

		// Jika selesai, tandai pembayaran terkait menjadi Selesai juga
		if tStatus == "Selesai" {
			pays, _ := repository.GetPembayaranFiltered(bson.M{"transaksi_id": existing.TransaksiID})
			for _, p := range pays {
				_, _ = repository.UpdatePembayaran(p.ID, models.Pembayaran{Status: "Selesai"})
			}
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
