package controllers

import (
	"backend/models"
	"backend/repository"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

// GET /kategori - semua role bisa melihat
func GetAllKategori(c *fiber.Ctx) error {
	list, err := repository.GetAllKategori()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal mengambil kategori"})
	}
	return c.JSON(list)
}

// GET /kategori/:id
func GetKategoriByID(c *fiber.Ctx) error {
	id := c.Params("id")
	k, err := repository.GetKategoriByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Kategori tidak ditemukan"})
	}
	return c.JSON(k)
}

// POST /kategori - hanya admin & gudang
func CreateKategori(c *fiber.Ctx) error {
	var input models.Kategori
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Data tidak valid"})
	}

	// Generate ID KTGxxx dari counters
	id, err := repository.GenerateID("kategori")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal generate ID"})
	}
	input.ID = id
	input.CreatedAt = time.Now()

	if _, err := repository.CreateKategori(&input); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal membuat kategori"})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Kategori berhasil dibuat", "id": input.ID})
}

// PUT /kategori/:id - hanya admin & gudang
func UpdateKategori(c *fiber.Ctx) error {
	id := c.Params("id")
	var body struct {
		NamaKategori string `json:"nama_kategori"`
		Deskripsi    string `json:"deskripsi"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Data tidak valid"})
	}
	update := bson.M{}
	if body.NamaKategori != "" {
		update["nama_kategori"] = body.NamaKategori
	}
	if body.Deskripsi != "" {
		update["deskripsi"] = body.Deskripsi
	}

	if len(update) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Tidak ada perubahan"})
	}
	if _, err := repository.UpdateKategori(id, update); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal mengupdate kategori"})
	}
	return c.JSON(fiber.Map{"message": "Kategori berhasil diupdate"})
}

// DELETE /kategori/:id - hanya admin & gudang
func DeleteKategori(c *fiber.Ctx) error {
	id := c.Params("id")
	if _, err := repository.DeleteKategori(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal menghapus kategori"})
	}
	return c.JSON(fiber.Map{"message": "Kategori berhasil dihapus"})
}
