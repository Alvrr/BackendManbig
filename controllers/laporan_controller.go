package controllers

import (
	"context"
	"fmt"
	"time"

	"backend/config"
	"backend/models"
	"backend/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type LaporanController struct {
	DB *mongo.Database
}

func NewLaporanController() *LaporanController {
	return &LaporanController{
		DB: config.DB,
	}
}

// BestSellers returns top selling products within last N days
func (lc *LaporanController) BestSellers(c *fiber.Ctx) error {
	daysParam := c.Query("days", "7")
	limitParam := c.Query("limit", "5")

	days := 7
	limit := 5
	if d, err := parseInt(daysParam); err == nil && d > 0 {
		days = d
	}
	if l, err := parseInt(limitParam); err == nil && l > 0 {
		limit = l
	}

	end := time.Now()
	start := end.AddDate(0, 0, -days)

	list, err := repository.GetBestSellers(start, end, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(list)
}

func parseInt(s string) (int, error) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}

func (lc *LaporanController) ExportExcel(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	f := excelize.NewFile()

	// ===============================
	// STYLE HEADER
	// ===============================
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
	})

	// ===============================
	// SHEET 1 : DETAIL PRODUK
	// ===============================
	sheetDetail := "Detail Produk"
	f.SetSheetName("Sheet1", sheetDetail)

	headersDetail := []string{
		"ID Transaksi",
		"Tanggal",
		"Pelanggan",
		"Kasir",
		"Driver",
		"Jenis Pengiriman",
		"Nama Produk",
		"Jumlah",
		"Harga",
		"Subtotal",
		"Ongkir",
		"Status",
	}

	for i, h := range headersDetail {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetDetail, cell, h)
		f.SetCellStyle(sheetDetail, cell, cell, headerStyle)
	}

	trxColl := lc.DB.Collection("transaksi")
	userColl := lc.DB.Collection("user")
	pelangganColl := lc.DB.Collection("pelanggan")
	pengirimanColl := lc.DB.Collection("pengiriman")

	cursor, err := trxColl.Find(ctx, bson.M{})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer cursor.Close(ctx)

	row := 2
	var totalSubtotal float64
	var totalOngkir float64
	seenTrx := make(map[string]bool)
	for cursor.Next(ctx) {
		var trx models.Transaksi
		cursor.Decode(&trx)

		// kasir
		var kasir models.User
		_ = userColl.FindOne(ctx, bson.M{"_id": trx.KasirID}).Decode(&kasir)

		// pelanggan
		var pelanggan models.Pelanggan
		_ = pelangganColl.FindOne(ctx, bson.M{"_id": trx.PelangganID}).Decode(&pelanggan)

		// pengiriman + driver
		var pengiriman models.Pengiriman
		var driver models.User
		_ = pengirimanColl.FindOne(ctx, bson.M{"transaksi_id": trx.ID}).Decode(&pengiriman)
		_ = userColl.FindOne(ctx, bson.M{"_id": pengiriman.DriverID}).Decode(&driver)

		for i, item := range trx.Items {
			subtotal := float64(item.Jumlah) * item.Harga
			// Akumulasi subtotal semua item
			totalSubtotal += subtotal
			// Akumulasi ongkir per transaksi (hanya sekali)
			if !seenTrx[trx.ID] {
				totalOngkir += pengiriman.Ongkir
				seenTrx[trx.ID] = true
			}

			values := []interface{}{
				trx.ID,
				trx.CreatedAt.Format("02-01-2006 15:04"),
				pelanggan.Nama,
				kasir.Nama,
				driver.Nama,
				pengiriman.Jenis,
				item.NamaProduk,
				item.Jumlah,
				item.Harga,
				subtotal,
				// Tampilkan ongkir hanya di baris pertama item untuk transaksi ini
				func() interface{} {
					if i == 0 {
						return pengiriman.Ongkir
					}
					return ""
				}(),
				trx.Status,
			}

			for i, v := range values {
				cell, _ := excelize.CoordinatesToCellName(i+1, row)
				f.SetCellValue(sheetDetail, cell, v)
			}
			row++
		}
	}

	// Tambah baris TOTAL di akhir untuk Subtotal dan Ongkir
	totalRow := row
	// Label TOTAL di kolom I (opsional)
	f.SetCellValue(sheetDetail, "I"+itoa(totalRow), "TOTAL")
	// Total Subtotal (kolom J)
	f.SetCellValue(sheetDetail, "J"+itoa(totalRow), totalSubtotal)
	// Total Ongkir (kolom K)
	f.SetCellValue(sheetDetail, "K"+itoa(totalRow), totalOngkir)

	f.AutoFilter(sheetDetail, "A1:L1", []excelize.AutoFilterOptions{})
	f.SetPanes(sheetDetail, &excelize.Panes{
		Freeze: true,
		Split:  true,
		YSplit: 1,
	})

	// ===============================
	// SHEET 2 : RINGKASAN TRANSAKSI
	// ===============================
	sheetRingkas := "Ringkasan Transaksi"
	f.NewSheet(sheetRingkas)

	headersRingkas := []string{
		"ID Transaksi",
		"Tanggal",
		"Pelanggan",
		"Total Produk",
		"Ongkir",
		"Total Bayar",
		"Status",
	}

	for i, h := range headersRingkas {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetRingkas, cell, h)
		f.SetCellStyle(sheetRingkas, cell, cell, headerStyle)
	}

	cursor2, _ := trxColl.Find(ctx, bson.M{})
	defer cursor2.Close(ctx)

	row = 2
	var ringkasTotalSubtotal float64
	var ringkasTotalOngkir float64
	for cursor2.Next(ctx) {
		var trx models.Transaksi
		cursor2.Decode(&trx)

		var pelanggan models.Pelanggan
		_ = pelangganColl.FindOne(ctx, bson.M{"_id": trx.PelangganID}).Decode(&pelanggan)

		var pengiriman models.Pengiriman
		_ = pengirimanColl.FindOne(ctx, bson.M{"transaksi_id": trx.ID}).Decode(&pengiriman)

		// Akumulasi total di sheet ringkasan
		ringkasTotalSubtotal += trx.TotalHarga
		ringkasTotalOngkir += pengiriman.Ongkir

		values := []interface{}{
			trx.ID,
			trx.CreatedAt.Format("02-01-2006"),
			pelanggan.Nama,
			trx.TotalProduk,
			pengiriman.Ongkir,
			trx.TotalHarga + pengiriman.Ongkir,
			trx.Status,
		}

		for i, v := range values {
			cell, _ := excelize.CoordinatesToCellName(i+1, row)
			f.SetCellValue(sheetRingkas, cell, v)
		}
		row++
	}

	// Tambah baris TOTAL di akhir untuk Ongkir dan Subtotal (Total Harga)
	totalRow2 := row
	// Label TOTAL di kolom C (opsional)
	f.SetCellValue(sheetRingkas, "C"+itoa(totalRow2), "TOTAL")
	// Total Ongkir (kolom E)
	f.SetCellValue(sheetRingkas, "E"+itoa(totalRow2), ringkasTotalOngkir)
	// Total Subtotal/Produk (kolom D atau Total Harga di kolom F?)
	// Isi total subtotal (Total Harga tanpa ongkir) di kolom D sebagai referensi jumlah nilai produk
	f.SetCellValue(sheetRingkas, "D"+itoa(totalRow2), ringkasTotalSubtotal)

	f.AutoFilter(sheetRingkas, "A1:G1", []excelize.AutoFilterOptions{})
	f.SetPanes(sheetRingkas, &excelize.Panes{
		Freeze: true,
		Split:  true,
		YSplit: 1,
	})

	// ===============================
	// RESPONSE
	// ===============================
	f.SetActiveSheet(0)

	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", "attachment; filename=laporan_mbg.xlsx")

	buf, _ := f.WriteToBuffer()
	return c.Send(buf.Bytes())
}

// helper
func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}
