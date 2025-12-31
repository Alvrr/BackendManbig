package controllers

import (
	"backend/models"
	"backend/repository"
	"time"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	excelize "github.com/xuri/excelize/v2"
)

// GET /stok/saldo/:produk_id (view semua role)
func GetSaldoProduk(c *fiber.Ctx) error {
	id := c.Params("produk_id")
	s, err := repository.GetSaldoProduk(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal menghitung saldo"})
	}
	return c.JSON(s)
}

// GET /stok/mutasi/:produk_id (view semua role)
func GetMutasiByProduk(c *fiber.Ctx) error {
	id := c.Params("produk_id")
	list, err := repository.GetMutasiByProduk(id, 0, 0)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal mengambil mutasi"})
	}
	return c.JSON(list)
}

// GET /stok/mutasi (global list with filters) - view semua role
func ListMutasi(c *fiber.Ctx) error {
	// Query params: produk_id, jenis, keterangan, ref_type, ref_id, start, end, page, page_size
	filter := bson.M{}
	produkID := c.Query("produk_id")
	jenis := c.Query("jenis")
	ket := c.Query("keterangan")
	refType := c.Query("ref_type")
	refID := c.Query("ref_id")
	if produkID != "" { filter["produk_id"] = produkID }
	if jenis != "" { filter["jenis"] = jenis }
	if ket != "" { filter["keterangan"] = ket }
	if refType != "" { filter["ref_type"] = refType }
	if refID != "" { filter["ref_id"] = refID }
	// Date range
	start := c.Query("start")
	end := c.Query("end")
	if start != "" || end != "" {
		// created_at between start and end
		rangeFilter := bson.M{}
		if start != "" {
			if t, err := time.Parse(time.RFC3339, start); err == nil {
				rangeFilter["$gte"] = t
			}
		}
		if end != "" {
			if t, err := time.Parse(time.RFC3339, end); err == nil {
				rangeFilter["$lte"] = t
			}
		}
		if len(rangeFilter) > 0 {
			filter["created_at"] = rangeFilter
		}
	}
	page, _ := strconv.Atoi(c.Query("page", "0"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "0"))
	sortDesc := c.Query("sort", "desc") != "asc"
	list, err := repository.ListMutasi(filter, page, pageSize, sortDesc)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal mengambil mutasi"})
	}
	return c.JSON(list)
}

// GET /stok/mutasi/export (admin, gudang) - export filtered mutasi to Excel
func ExportMutasiExcel(c *fiber.Ctx) error {
	// Reuse same filters as ListMutasi
	filter := bson.M{}
	produkID := c.Query("produk_id")
	jenis := c.Query("jenis")
	ket := c.Query("keterangan")
	refType := c.Query("ref_type")
	refID := c.Query("ref_id")
	if produkID != "" { filter["produk_id"] = produkID }
	if jenis != "" { filter["jenis"] = jenis }
	if ket != "" { filter["keterangan"] = ket }
	if refType != "" { filter["ref_type"] = refType }
	if refID != "" { filter["ref_id"] = refID }
	start := c.Query("start")
	end := c.Query("end")
	if start != "" || end != "" {
		rangeFilter := bson.M{}
		if start != "" {
			if t, err := time.Parse(time.RFC3339, start); err == nil {
				rangeFilter["$gte"] = t
			}
		}
		if end != "" {
			if t, err := time.Parse(time.RFC3339, end); err == nil {
				rangeFilter["$lte"] = t
			}
		}
		if len(rangeFilter) > 0 {
			filter["created_at"] = rangeFilter
		}
	}
	list, err := repository.ListMutasi(filter, 0, 0, true)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal mengambil mutasi"})
	}

	f := excelize.NewFile()
	sheet := f.GetSheetName(f.GetActiveSheetIndex())
	headers := []string{"Tanggal", "Produk ID", "Jenis", "Jumlah", "User ID", "Ref ID", "Ref Type", "Keterangan"}
	for i, h := range headers {
		col, _ := excelize.ColumnNumberToName(i + 1)
		_ = f.SetCellValue(sheet, col+"1", h)
	}
	for idx, m := range list {
		row := idx + 2
		_ = f.SetCellValue(sheet, "A"+strconv.Itoa(row), m.CreatedAt.Format(time.RFC3339))
		_ = f.SetCellValue(sheet, "B"+strconv.Itoa(row), m.ProdukID)
		_ = f.SetCellValue(sheet, "C"+strconv.Itoa(row), m.Jenis)
		_ = f.SetCellValue(sheet, "D"+strconv.Itoa(row), m.Jumlah)
		_ = f.SetCellValue(sheet, "E"+strconv.Itoa(row), m.UserID)
		_ = f.SetCellValue(sheet, "F"+strconv.Itoa(row), m.RefID)
		_ = f.SetCellValue(sheet, "G"+strconv.Itoa(row), m.RefType)
		_ = f.SetCellValue(sheet, "H"+strconv.Itoa(row), m.Keterangan)
	}
	// Prepare download
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", "attachment; filename=riwayat_stok.xlsx")
	if err := f.Write(c.Response().BodyWriter()); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal menulis file"})
	}
	return nil
}

// POST /stok (admin+gudang)
func CreateMutasi(c *fiber.Ctx) error {
	var m models.StokMutasi
	if err := c.BodyParser(&m); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Data tidak valid"})
	}
	if m.ProdukID == "" || m.Jenis == "" || m.Jumlah == 0 {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"message": "produk_id, jenis, jumlah wajib"})
	}
	// Normalisasi untuk adjust: ubah menjadi delta masuk/keluar dibanding saldo saat ini
	if m.Jenis == "adjust" {
		saldo, err := repository.GetSaldoProduk(m.ProdukID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal membaca saldo untuk adjust"})
		}
		// Jika sama, tidak perlu perubahan; jadikan jumlah 0 masuk agar tidak mempengaruhi saldo
		if m.Jumlah == saldo.Saldo {
			m.Jenis = "masuk"
			m.Jumlah = 0
		} else if m.Jumlah > saldo.Saldo {
			m.Jenis = "masuk"
			m.Jumlah = m.Jumlah - saldo.Saldo
		} else {
			m.Jenis = "keluar"
			m.Jumlah = saldo.Saldo - m.Jumlah
		}
	}
	// Set ID dari counter stok + created_at
	id, err := repository.GenerateID("stok")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Gagal generate ID"})
	}
	m.ID = id
	       // Wajib isi user_id dari token, jika tidak ada tolak
	       uid, ok := c.Locals("userID").(string)
	       if !ok || uid == "" {
		       return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "User login tidak valid"})
	       }
	       m.UserID = uid
	       // Mutasi manual: ref_type harus 'manual', ref_id dikosongkan
	       if m.RefType == "" || m.RefType == "manual" {
		       m.RefType = "manual"
		       m.RefID = ""
	       }
	       m.CreatedAt = time.Now()
	       if _, err := repository.CreateMutasi(&m); err != nil {
		       return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	       }
	       return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Mutasi stok berhasil dibuat", "id": m.ID})
}
