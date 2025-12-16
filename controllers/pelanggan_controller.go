package controllers

import (
	"backend/models"
	"backend/repository"
	"backend/utils"

	"github.com/gofiber/fiber/v2"
)

// GetAllPelanggan godoc
//
//	@Summary		Get all customers
//	@Description	Mengambil semua data pelanggan
//	@Tags			Pelanggan
//	@Security		BearerAuth
//	@Produce		json
//	@Success		200	{array}		models.Pelanggan
//	@Failure		500	{object}	map[string]interface{}	"Internal Server Error"
//	@Router			/pelanggan [get]
func GetAllPelanggan(c *fiber.Ctx) error {
	data, err := repository.GetAllPelanggan()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Gagal mengambil data pelanggan",
			"error":   err.Error(),
		})
	}
	return c.JSON(data)
}

// GetPelangganByID godoc
//
//	@Summary		Get customer by ID
//	@Description	Mengambil data pelanggan berdasarkan ID
//	@Tags			Pelanggan
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		string	true	"Customer ID"
//	@Success		200	{object}	models.Pelanggan
//	@Failure		404	{object}	map[string]interface{}	"Pelanggan tidak ditemukan"
//	@Router			/pelanggan/{id} [get]
func GetPelangganByID(c *fiber.Ctx) error {
	id := c.Params("id")
	pelanggan, err := repository.GetPelangganByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Pelanggan tidak ditemukan",
			"error":   err.Error(),
		})
	}
	return c.JSON(pelanggan)
}

// CreatePelanggan godoc
//
//	@Summary		Create customer
//	@Description	Membuat pelanggan baru
//	@Tags			Pelanggan
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			pelanggan	body		models.Pelanggan		true	"Customer data"
//	@Success		201			{object}	map[string]interface{}	"Pelanggan berhasil ditambahkan"
//	@Failure		400			{object}	map[string]interface{}	"Request tidak valid"
//	@Failure		422			{object}	map[string]interface{}	"Validasi gagal"
//	@Router			/pelanggan [post]
func CreatePelanggan(c *fiber.Ctx) error {
	var pelanggan models.Pelanggan

	if err := c.BodyParser(&pelanggan); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Request tidak valid",
			"error":   err.Error(),
		})
	}

	// ✅ Validasi input
	if err := utils.Validate.Struct(pelanggan); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"message": "Validasi gagal",
			"error":   err.Error(),
		})
	}

	newID, err := repository.GenerateID("pelanggan")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Gagal generate ID pelanggan",
			"error":   err.Error(),
		})
	}

	pelanggan.ID = newID

	result, err := repository.CreatePelanggan(pelanggan)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Gagal menambahkan pelanggan",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Pelanggan berhasil ditambahkan",
		"data":    result.InsertedID,
	})
}

// UpdatePelanggan godoc
//
//	@Summary		Update customer
//	@Description	Update data pelanggan berdasarkan ID
//	@Tags			Pelanggan
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			id			path		string					true	"Customer ID"
//	@Param			pelanggan	body		models.Pelanggan		true	"Customer data"
//	@Success		200			{object}	map[string]interface{}	"Pelanggan berhasil diupdate"
//	@Failure		400			{object}	map[string]interface{}	"Request tidak valid"
//	@Failure		422			{object}	map[string]interface{}	"Validasi gagal"
//	@Router			/pelanggan/{id} [put]
func UpdatePelanggan(c *fiber.Ctx) error {
	id := c.Params("id")
	var pelanggan models.Pelanggan

	if err := c.BodyParser(&pelanggan); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Request tidak valid",
			"error":   err.Error(),
		})
	}

	// ✅ Validasi input - pastikan field required tidak kosong
	if pelanggan.Nama == "" || pelanggan.Email == "" || pelanggan.NoHP == "" || pelanggan.Alamat == "" {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"message": "Validasi gagal",
			"error":   "Nama, email, no HP, dan alamat wajib diisi",
		})
	}

	_, err := repository.UpdatePelanggan(id, pelanggan)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Gagal update pelanggan",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Pelanggan berhasil diupdate",
	})
}

// DeletePelanggan godoc
//
//	@Summary		Delete customer
//	@Description	Hapus pelanggan berdasarkan ID
//	@Tags			Pelanggan
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		string					true	"Customer ID"
//	@Success		200	{object}	map[string]interface{}	"Pelanggan berhasil dihapus"
//	@Failure		500	{object}	map[string]interface{}	"Internal Server Error"
//	@Router			/pelanggan/{id} [delete]
func DeletePelanggan(c *fiber.Ctx) error {
	id := c.Params("id")

	_, err := repository.DeletePelanggan(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Gagal hapus pelanggan",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Pelanggan berhasil dihapus",
	})
}
