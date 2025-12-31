package controllers

import (
	"backend/models"
	"backend/repository"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
)

// GetAllPembayaran godoc
//
//	@Summary		Get all payments
//	@Description	Mengambil semua data pembayaran berdasarkan role user
//	@Tags			Pembayaran
//	@Security		BearerAuth
//	@Produce		json
//	@Success		200	{array}		models.Pembayaran
//	@Failure		500	{object}	map[string]interface{}	"Internal Server Error"
//	@Router			/pembayaran [get]
func GetAllPembayaran(c *fiber.Ctx) error {
	role := c.Locals("userRole").(string)
	id := c.Locals("userID").(string)

	filter := bson.M{}
	if role != "admin" {
		filter["kasir_id"] = id
	}

	data, err := repository.GetPembayaranFiltered(filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Gagal ambil data pembayaran",
			"error":   err.Error(),
		})
	}
	return c.JSON(data)
}

// GetPembayaranByID godoc
//
//	@Summary		Get payment by ID
//	@Description	Mengambil data pembayaran berdasarkan ID
//	@Tags			Pembayaran
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		string	true	"Payment ID"
//	@Success		200	{object}	models.Pembayaran
//	@Failure		404	{object}	map[string]interface{}	"Data tidak ditemukan"
//	@Router			/pembayaran/{id} [get]
func GetPembayaranByID(c *fiber.Ctx) error {
	id := c.Params("id")
	role := c.Locals("userRole").(string)
	userID := c.Locals("userID").(string)

	data, err := repository.GetPembayaranByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Data tidak ditemukan",
			"error":   err.Error(),
		})
	}

	if role != "admin" && data.KasirID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Akses ditolak"})
	}

	return c.JSON(data)
}

// ID dihasilkan via counters repository dengan nama "pembayaran"

// CreatePembayaran godoc
//
//	@Summary		Create payment
//	@Description	Membuat pembayaran baru
//	@Tags			Pembayaran
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			pembayaran	body		models.Pembayaran		true	"Payment data"
//	@Success		201			{object}	map[string]interface{}	"Pembayaran berhasil dibuat"
//	@Failure		400			{object}	map[string]interface{}	"Request tidak valid"
//	@Failure		422			{object}	map[string]interface{}	"Validasi gagal"
//	@Router			/pembayaran [post]
func CreatePembayaran(c *fiber.Ctx) error {
	var pembayaran models.Pembayaran
	if err := c.BodyParser(&pembayaran); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Request tidak valid",
			"error":   err.Error(),
		})
	}
	// Set kasir dari token
	pembayaran.KasirID = c.Locals("userID").(string)

	// Validasi minimal untuk skema baru
	if pembayaran.TransaksiID == "" || pembayaran.TotalBayar <= 0 {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"message": "transaksi_id dan total_bayar wajib"})
	}

	id, err := repository.GenerateID("pembayaran")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Gagal generate ID",
			"error":   err.Error(),
		})
	}
	pembayaran.ID = id
	pembayaran.CreatedAt = time.Now()
	if pembayaran.Status == "" {
		pembayaran.Status = "pending"
	}

	result, err := repository.CreatePembayaran(pembayaran)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Gagal simpan data",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Berhasil ditambahkan",
		"data":    result.InsertedID,
	})
}

// SelesaikanPembayaran godoc
//
//	@Summary		Complete payment
//	@Description	Menyelesaikan pembayaran (ubah status menjadi selesai)
//	@Tags			Pembayaran
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		string					true	"Payment ID"
//	@Success		200	{object}	map[string]interface{}	"Transaksi berhasil diselesaikan"
//	@Failure		400	{object}	map[string]interface{}	"Transaksi sudah selesai"
//	@Failure		403	{object}	map[string]interface{}	"Akses ditolak"
//	@Failure		404	{object}	map[string]interface{}	"Data tidak ditemukan"
//	@Router			/pembayaran/selesaikan/{id} [put]
func SelesaikanPembayaran(c *fiber.Ctx) error {
	id := c.Params("id")
	role := c.Locals("userRole").(string)
	userID := c.Locals("userID").(string)

	pembayaran, err := repository.GetPembayaranByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Data tidak ditemukan",
		})
	}

	if role != "admin" && pembayaran.KasirID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Akses ditolak"})
	}

	if pembayaran.Status == "Selesai" || pembayaran.Status == "selesai" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Transaksi sudah selesai",
		})
	}

	err = repository.SelesaikanPembayaran(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Gagal menyelesaikan",
			"error":   err.Error(),
		})
	}
	// Tandai transaksi terkait menjadi Selesai
	if pembayaran.TransaksiID != "" {
		_, _ = repository.UpdateTransaksi(pembayaran.TransaksiID, bson.M{"status": "Selesai"})
		// Mutasi stok yang berelasi dengan transaksi tersebut diberi keterangan 'terjual'
		_ = repository.UpdateMutasiKeteranganByRef(pembayaran.TransaksiID, "terjual")
	}
	return c.JSON(fiber.Map{
		"message": "Transaksi berhasil diselesaikan",
	})
}

// CetakSuratJalan dihapus dari pembayaran; akan dikelola oleh modul pengiriman.
