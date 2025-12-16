package controllers

import (
	"backend/models"
	"backend/repository"
	"backend/utils"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// Register function (internal use only, tidak ditampilkan di Swagger)
func Register(c *fiber.Ctx) error {
	var input models.User

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Data tidak valid"})
	}

	// Validasi role wajib (admin/kasir/gudang/driver)
	role := strings.ToLower(input.Role)
	if role != "admin" && role != "kasir" && role != "gudang" && role != "driver" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Role harus admin, kasir, gudang, atau driver"})
	}

	// Generate ID otomatis dari counters.go (jangan ambil dari frontend)
	id, err := repository.GenerateID(role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal membuat ID user"})
	}
	input.ID = id

	// Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), 12)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal enkripsi password"})
	}
	input.Password = string(hashed)
	input.CreatedAt = time.Now()

	// Simpan user
	if err := repository.CreateUser(&input); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal registrasi user"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Registrasi berhasil",
		"id":      input.ID,
	})
}

// Login godoc
//
//	@Summary		Login user
//	@Description	Login user dengan email dan password
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			user	body		models.LoginInput		true	"Login credentials"
//	@Success		200		{object}	map[string]interface{}	"Login berhasil"
//	@Failure		400		{object}	map[string]interface{}	"Request tidak valid"
//	@Failure		401		{object}	map[string]interface{}	"Email atau password salah"
//	@Router			/auth/login [post]
func Login(c *fiber.Ctx) error {
	var input models.LoginInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Data tidak valid"})
	}

	user, err := repository.FindUserByEmail(input.Email)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Email atau password salah"})
	}

	// Cek status aktif/nonaktif
	if user.Status != "aktif" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Akun karyawan nonaktif, tidak dapat login"})
	}

	// Bandingkan password plaintext dan hashed
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Email atau password salah"})
	}

	// ✅ Perbaikan: tambahkan user.Nama
	token, err := utils.GenerateToken(user.ID, user.Role, user.Nama)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal membuat token"})
	}

	return c.JSON(fiber.Map{
		"message": "Login berhasil",
		"token":   token,
	})
}
