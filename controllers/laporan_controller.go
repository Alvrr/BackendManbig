package controllers

import (
	"backend/config"
	"backend/models"
	"bytes"
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson"
)

// ExportLaporanExcel godoc
//
//	@Summary		Export laporan ke Excel
//	@Description	Export semua data pembayaran ke file Excel
//	@Tags			Laporan
//	@Security		BearerAuth
//	@Produce		application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
//	@Success		200	{file}		binary					"File Excel berhasil diexport"
//	@Failure		500	{object}	map[string]interface{}	"Internal Server Error"
//	@Router			/laporan/excel [get]
func ExportLaporanExcel(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := config.PembayaranCollection.Find(ctx, bson.M{})
	if err != nil {
		return c.Status(500).SendString("Gagal mengambil data")
	}
	defer cursor.Close(ctx)

	f := excelize.NewFile()
	sheet := "Laporan"
	f.SetSheetName("Sheet1", sheet)

	// Header disesuaikan dengan skema pembayaran baru
	headers := []string{"ID Pembayaran", "ID Transaksi", "Nama Kasir", "Metode", "Total Bayar", "Status", "Tanggal"}
	columns := []string{"A", "B", "C", "D", "E", "F", "G"}

	// ✅ Buat style header (bold + center + background)
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#f2f2f2"},
			Pattern: 1,
		},
	})

	for i, h := range headers {
		cell := columns[i] + "1"
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
	}

	// Data
	row := 2
	for cursor.Next(ctx) {
		var bayar models.Pembayaran
		if err := cursor.Decode(&bayar); err != nil {
			continue
		}

		// Ambil nama kasir dari user collection
		kasirNama := "Tidak ditemukan"
		var kasir models.User
		if err := config.UserCollection.FindOne(ctx, bson.M{"_id": bayar.KasirID}).Decode(&kasir); err == nil {
			kasirNama = kasir.Nama
		}

		values := []interface{}{
			bayar.ID,
			bayar.TransaksiID,
			kasirNama,
			bayar.Metode,
			bayar.TotalBayar,
			bayar.Status,
			bayar.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		for col, val := range values {
			cell, _ := excelize.CoordinatesToCellName(col+1, row)
			f.SetCellValue(sheet, cell, val)
		}
		row++
	}

	// Lebar kolom otomatis
	for _, col := range columns {
		f.SetColWidth(sheet, col, col, 25)
	}

	// Output Excel ke browser
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return c.Status(500).SendString("Gagal generate Excel")
	}

	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", "attachment;filename=laporan.xlsx")
	return c.SendStream(&buf)
}
