package controllers

import (
	"backend/repository"

	"github.com/gofiber/fiber/v2"
)

// GetRiwayatPembayaran godoc
//	@Summary		Get payment history
//	@Description	Mengambil riwayat pembayaran berdasarkan role user
//	@Tags			Riwayat
//	@Security		BearerAuth
//	@Produce		json
//	@Success		200	{array}		models.Pembayaran
//	@Failure		500	{object}	map[string]interface{}	"Internal Server Error"
//	@Router			/riwayat [get]
func GetRiwayatPembayaran(c *fiber.Ctx) error {
	role := c.Locals("userRole").(string)
	id := c.Locals("userID").(string)
	filter := map[string]interface{}{}
	if role == "driver" {
		filter["id_driver"] = id
	} else if role == "kasir" {
		filter["id_kasir"] = id
	}
	data, err := repository.GetRiwayatPembayaran(filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Gagal ambil riwayat pembayaran",
			"error":   err.Error(),
		})
	}
	return c.JSON(data)
}
