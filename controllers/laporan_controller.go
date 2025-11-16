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

	// Header
	headers := []string{"ID Transaksi", "Nama Pembeli", "Nama Kasir", "Nama Driver", "Produk", "Qty", "Harga", "Subtotal", "Tanggal"}
	columns := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I"}

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

		// Ambil nama pelanggan
		var pelanggan models.Pelanggan
		err := config.PelangganCollection.FindOne(ctx, bson.M{"_id": bayar.IDPelanggan}).Decode(&pelanggan)
		if err != nil {
			pelanggan.Nama = "Tidak ditemukan"
		}

		// Untuk setiap produk, buat baris sendiri
		for _, item := range bayar.Produk {
			values := []interface{}{
				bayar.ID,
				pelanggan.Nama,
				bayar.NamaKasir,
				bayar.NamaDriver,
				item.NamaProduk,
				item.Jumlah,
				item.Harga,
				item.Subtotal,
				bayar.Tanggal,
			}
			for col, val := range values {
				cell, _ := excelize.CoordinatesToCellName(col+1, row)
				f.SetCellValue(sheet, cell, val)
			}
			row++
		}
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
