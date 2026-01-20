package controllers

import (
	"backend/models"
	"backend/repository"
	"strings"
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
	if role != "admin" && role != "kasir" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Akses ditolak"})
	}

	filter := bson.M{}
	if role != "admin" {
		// IMPORTANT: kasir hanya boleh melihat pembayaran miliknya sendiri
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
	if role != "admin" && role != "kasir" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Akses ditolak"})
	}

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
	role := c.Locals("userRole").(string)
	userID := c.Locals("userID").(string)
	if role != "kasir" {
		// Admin read-only; driver/gudang ditolak
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Akses ditolak"})
	}

	var body struct {
		TransaksiID    string `json:"transaksi_id"`
		Metode         string `json:"metode"`
		Delivery       bool   `json:"delivery"`
		JenisKendaraan string `json:"jenis_kendaraan"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Request tidak valid",
			"error":   err.Error(),
		})
	}

	if body.TransaksiID == "" {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"message": "transaksi_id wajib"})
	}
	metode := strings.TrimSpace(body.Metode)
	if metode == "" {
		metode = "cash"
	}

	// Ownership check: transaksi harus milik kasir login
	trx, err := repository.GetTransaksiByID(body.TransaksiID)
	if err != nil || trx == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Transaksi tidak ditemukan"})
	}
	if trx.KasirID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Akses ditolak"})
	}

	// Hitung total_toko dari transaksi (server-side)
	totalToko := trx.TotalHarga
	if totalToko <= 0 {
		var sum float64
		for _, it := range trx.Items {
			sum += float64(it.Jumlah) * it.Harga
		}
		totalToko = sum
	}

	// Hitung ongkir dari kendaraan (server-side) jika delivery
	ongkir := float64(0)
	if body.Delivery {
		jenis := strings.ToLower(strings.TrimSpace(body.JenisKendaraan))
		if jenis == "" {
			jenis = "mobil"
		}
		if jenis == "mobil" {
			ongkir = 25000
		} else if jenis == "motor" {
			ongkir = 10000
		} else {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"message": "jenis_kendaraan harus salah satu dari: mobil, motor"})
		}
	}

	// Build pembayaran final (semua nilai finansial dihitung server)
	pembayaran := models.Pembayaran{
		TransaksiID: body.TransaksiID,
		KasirID:     userID,
		Metode:      metode,
		TotalBayar:  totalToko + ongkir,
		Status:      "pending",
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
	if role != "kasir" {
		// Admin read-only; driver/gudang ditolak
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Akses ditolak"})
	}

	pembayaran, err := repository.GetPembayaranByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Data tidak ditemukan",
		})
	}

	if pembayaran.KasirID != userID {
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
