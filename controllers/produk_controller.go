package controllers

import (
	"backend/models"
	"backend/repository"
	"time"

	"github.com/gofiber/fiber/v2"
)

// GetAllProduk godoc
//
//	@Summary		Get all products
//	@Description	Mengambil semua data produk
//	@Tags			Produk
//	@Security		BearerAuth
//	@Produce		json
//	@Success		200	{array}		models.ProdukSwagger
//	@Failure		500	{object}	map[string]interface{}	"Internal Server Error"
//	@Router			/produk [get]
func GetAllProduk(c *fiber.Ctx) error {
	produks, err := repository.GetAllProduk()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Gagal mengambil data produk",
			"error":   err.Error(),
		})
	}
	return c.JSON(produks)
}

// GetProdukByID godoc
//
//	@Summary		Get product by ID
//	@Description	Mengambil data produk berdasarkan ID
//	@Tags			Produk
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		string	true	"Product ID"
//	@Success		200	{object}	models.ProdukSwagger
//	@Failure		404	{object}	map[string]interface{}	"Produk tidak ditemukan"
//	@Router			/produk/{id} [get]
func GetProdukByID(c *fiber.Ctx) error {
	id := c.Params("id")
	produk, err := repository.GetProdukByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Produk tidak ditemukan",
			"error":   err.Error(),
		})
	}
	return c.JSON(produk)
}

// CreateProduk godoc
//
//	@Summary		Create product
//	@Description	Membuat produk baru
//	@Tags			Produk
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			produk	body		models.ProdukInput		true	"Product data"
//	@Success		201		{object}	map[string]interface{}	"Produk berhasil ditambahkan"
//	@Failure		400		{object}	map[string]interface{}	"Request tidak valid"
//	@Failure		422		{object}	map[string]interface{}	"Validasi gagal"
//	@Router			/produk [post]
func CreateProduk(c *fiber.Ctx) error {
	var produk models.Produk

	if err := c.BodyParser(&produk); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Request tidak valid",
			"error":   err.Error(),
		})
	}

	// Basic validation from provided schema
	if produk.NamaProduk == "" || produk.KategoriID == "" || produk.HargaJual <= 0 {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"message": "Validasi gagal",
			"error":   "nama_produk, kategori_id, harga_jual wajib diisi",
		})
	}

	// 🔢 Generate ID dan waktu
	newID, err := repository.GenerateID("produk")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Gagal generate ID produk",
			"error":   err.Error(),
		})
	}

	produk.ID = newID
	produk.CreatedAt = time.Now()

	result, err := repository.CreateProduk(produk)
	if err != nil {
		status := fiber.StatusInternalServerError
		msg := err.Error()
		if msg == "kategori tidak ditemukan" {
			status = fiber.StatusUnprocessableEntity
		}
		return c.Status(status).JSON(fiber.Map{
			"message": "Gagal menambahkan produk",
			"error":   msg,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Produk berhasil ditambahkan",
		"data":    result.InsertedID,
	})
}

// UpdateProduk godoc
//
//	@Summary		Update product
//	@Description	Update data produk berdasarkan ID
//	@Tags			Produk
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"Product ID"
//	@Param			produk	body		models.ProdukInput		true	"Product data"
//	@Success		200		{object}	map[string]interface{}	"Produk berhasil diupdate"
//	@Failure		400		{object}	map[string]interface{}	"Request tidak valid"
//	@Failure		422		{object}	map[string]interface{}	"Validasi gagal"
//	@Router			/produk/{id} [put]
func UpdateProduk(c *fiber.Ctx) error {
	id := c.Params("id")
	var produk models.Produk

	if err := c.BodyParser(&produk); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Request tidak valid",
			"error":   err.Error(),
		})
	}

	// ✅ Validasi input - pastikan field required tidak kosong
	if produk.NamaProduk == "" || produk.KategoriID == "" || produk.HargaJual <= 0 || produk.Stok < 0 {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
			"message": "Validasi gagal",
			"error":   "Nama produk, kategori_id, harga_jual (>0), dan stok (>=0) wajib diisi",
		})
	}

	_, err := repository.UpdateProduk(id, produk)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Gagal update produk",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Produk berhasil diupdate",
	})
}

// DeleteProduk godoc
//
//	@Summary		Delete product
//	@Description	Hapus produk berdasarkan ID
//	@Tags			Produk
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		string					true	"Product ID"
//	@Success		200	{object}	map[string]interface{}	"Produk berhasil dihapus"
//	@Failure		500	{object}	map[string]interface{}	"Internal Server Error"
//	@Router			/produk/{id} [delete]
func DeleteProduk(c *fiber.Ctx) error {
	id := c.Params("id")

	_, err := repository.DeleteProduk(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Gagal hapus produk",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Produk berhasil dihapus",
	})
}
