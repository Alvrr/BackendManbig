package main

import (
	"backend/config"
	_ "backend/docs" // Import docs for swagger
	"backend/middleware"
	"backend/repository"
	"backend/routes"
	"log"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

//	@title			Backend MBG API
//	@version		1.0
//	@description	API documentation untuk Backend GROSIR
//	@description
//	@description	**Sistem Login:**
//	@description	- Admin: restu129@gmail.com / restu123
//	@description	- Kasir: kasir129@gmail.com / 123456
//	@description	- Driver: driver129@gmail.com / 123456
//	@description
//	@description	**Authentication:**
//	@description	- Semua endpoint (kecuali login) memerlukan Bearer Token
//	@description	- Token didapat dari endpoint /auth/login
//	@description	- Format: Authorization: Bearer {token}
//	@description
//	@description	**Role Permissions:**
//	@description	- Admin: Akses penuh ke semua fitur
//	@description	- Kasir: CRUD produk, pelanggan, pembayaran, lihat riwayat
//	@description	- Driver: Lihat pembayaran assigned, update status selesai, cetak surat jalan
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host		backendmbg-production.up.railway.app
//	@BasePath	/
//	@schemes	https

//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.

func main() {
	// Load file .env (tidak fatal jika gagal, agar bisa jalan di Railway)
	_ = godotenv.Load()

	// Ensure JWT_SECRET in production; allow safe default in development
	appEnv := strings.ToLower(os.Getenv("APP_ENV"))
	if os.Getenv("JWT_SECRET") == "" {
		if appEnv == "production" {
			log.Fatal("❌ JWT_SECRET harus diset di environment production")
		}
		os.Setenv("JWT_SECRET", "dev_secret_key_change_me")
		log.Println("⚠️ JWT_SECRET tidak diset, menggunakan default untuk development")
	}

	// Koneksi ke MongoDB
	config.ConnectDB()

	// Inisialisasi counters yang diperlukan
	if err := repository.InitializeCounters(); err != nil {
		log.Printf("⚠️ Peringatan: %v", err)
	} else {
		log.Println("✅ Counters berhasil diinisialisasi")
	}

	// Pastikan index kategori (unique nama)
	if err := repository.EnsureKategoriIndexes(); err != nil {
		log.Printf("⚠️ Gagal membuat index kategori: %v", err)
	}

	// Pastikan index produk
	if err := repository.EnsureProdukIndexes(); err != nil {
		log.Printf("⚠️ Gagal membuat index produk: %v", err)
	}

	// Pastikan index stok
	if err := repository.EnsureStokIndexes(); err != nil {
		log.Printf("⚠️ Gagal membuat index stok: %v", err)
	}

	// Pastikan index transaksi
	if err := repository.EnsureTransaksiIndexes(); err != nil {
		log.Printf("⚠️ Gagal membuat index transaksi: %v", err)
	}

	// Pastikan index pembayaran
	if err := repository.EnsurePembayaranIndexes(); err != nil {
		log.Printf("⚠️ Gagal membuat index pembayaran: %v", err)
	}

	// Pastikan index pengiriman
	if err := repository.EnsurePengirimanIndexes(); err != nil {
		log.Printf("⚠️ Gagal membuat index pengiriman: %v", err)
	}

	// Pastikan index user (unique email & nama)
	if err := repository.EnsureUserIndexes(); err != nil {
		log.Printf("⚠️ Gagal membuat index user: %v", err)
	}

	// Inisialisasi Fiber
	app := fiber.New()

	// Middleware global
	app.Use(middleware.LoggerMiddleware())
	app.Use(middleware.CorsMiddleware())

	// JWTMiddleware global, kecuali untuk /auth/login dan /auth/register
	app.Use(func(c *fiber.Ctx) error {
		path := c.Path()
		// NOTE: export endpoints are opened via window.open and may pass token via query.
		// They are protected at route-level via JWTMiddlewareForExport + RoleGuard.
		if path == "/laporan/export/excel" {
			return c.Next()
		}
		if path == "/auth/login" || path == "/auth/register" || strings.HasPrefix(path, "/swagger") {
			return c.Next()
		}
		return middleware.JWTMiddleware(c)
	})

	// Swagger route
	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	// Semua route (termasuk auth/login/register)
	routes.SetupRoutes(app)

	// Port server (default ke 5000 agar konsisten dengan frontend & docs)
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	log.Println("🚀 Server jalan di http://localhost:" + port)
	log.Fatal(app.Listen(":" + port))
}
