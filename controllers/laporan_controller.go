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
	"go.mongodb.org/mongo-driver/mongo/options"
)

func buildCreatedAtFilterFromQuery(c *fiber.Ctx) (bson.M, error) {
	startStr := c.Query("start", "")
	endStr := c.Query("end", "")
	monthStr := c.Query("month", "")
	yearStr := c.Query("year", "")

	// Priority:
	// 1) explicit start/end
	// 2) month+year
	// 3) year
	if startStr != "" || endStr != "" {
		if startStr == "" || endStr == "" {
			return bson.M{}, fmt.Errorf("start dan end harus diisi bersamaan")
		}
		// UI memakai input tanggal lokal (type=date). Samakan dengan server: interpretasi sebagai waktu lokal.
		startDate, errStart := time.ParseInLocation("2006-01-02", startStr, time.Local)
		endDate, errEnd := time.ParseInLocation("2006-01-02", endStr, time.Local)
		if errStart != nil || errEnd != nil {
			return bson.M{}, fmt.Errorf("format tanggal harus YYYY-MM-DD")
		}
		return bson.M{
			"created_at": bson.M{
				"$gte": startDate,
				"$lt":  endDate.AddDate(0, 0, 1),
			},
		}, nil
	}

	if monthStr != "" {
		if yearStr == "" {
			return bson.M{}, fmt.Errorf("year wajib diisi jika month digunakan")
		}
		var year int
		if _, err := fmt.Sscanf(yearStr, "%d", &year); err != nil || year < 1900 {
			return bson.M{}, fmt.Errorf("year tidak valid")
		}
		var month int
		if _, err := fmt.Sscanf(monthStr, "%d", &month); err != nil || month < 1 || month > 12 {
			return bson.M{}, fmt.Errorf("month tidak valid")
		}
		startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
		endExclusive := startDate.AddDate(0, 1, 0)
		return bson.M{
			"created_at": bson.M{
				"$gte": startDate,
				"$lt":  endExclusive,
			},
		}, nil
	}

	if yearStr != "" {
		var year int
		if _, err := fmt.Sscanf(yearStr, "%d", &year); err != nil || year < 1900 {
			return bson.M{}, fmt.Errorf("year tidak valid")
		}
		startDate := time.Date(year, time.January, 1, 0, 0, 0, 0, time.Local)
		endExclusive := startDate.AddDate(1, 0, 0)
		return bson.M{
			"created_at": bson.M{
				"$gte": startDate,
				"$lt":  endExclusive,
			},
		}, nil
	}

	return bson.M{}, nil
}

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
	role, _ := c.Locals("userRole").(string)
	userID, _ := c.Locals("userID").(string)

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
	// IMPORTANT: dashboard kasir berbasis data miliknya sendiri
	// (field DB: kasir_id  diperlakukan sebagai created_by)
	kasirID := ""
	if role == "kasir" {
		kasirID = userID
	}

	list, err := repository.GetBestSellers(start, end, limit, kasirID)
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
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// ===============================
	// FILTERS (match UI: pembayaran-centric)
	// ===============================
	dateFilter, err := buildCreatedAtFilterFromQuery(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	filter := bson.M{}
	for k, v := range dateFilter {
		filter[k] = v
	}

	// UI excludes these pembayaran IDs
	excludeIDs := []string{"PMB005", "PMB004", "PMB003", "PMB002", "PMB001"}
	filter["_id"] = bson.M{"$nin": excludeIDs}

	// Export only Pending/Selesai (case defensive)
	filter["status"] = bson.M{"$in": []string{"pending", "selesai", "Pending", "Selesai"}}

	// ===============================
	// LOAD DATA (pembayaran as source of truth)
	// ===============================
	pembayaranColl := lc.DB.Collection("pembayaran")
	trxColl := lc.DB.Collection("transaksi")
	userColl := lc.DB.Collection("user")
	pelangganColl := lc.DB.Collection("pelanggan")
	pengirimanColl := lc.DB.Collection("pengiriman")
	produkColl := lc.DB.Collection("produk")

	findOpts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	curPay, err := pembayaranColl.Find(ctx, filter, findOpts)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	defer curPay.Close(ctx)

	payments := make([]models.Pembayaran, 0)
	trxIDSet := map[string]struct{}{}
	for curPay.Next(ctx) {
		var p models.Pembayaran
		if err := curPay.Decode(&p); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "gagal decode pembayaran"})
		}
		payments = append(payments, p)
		if p.TransaksiID != "" {
			trxIDSet[p.TransaksiID] = struct{}{}
		}
	}
	if err := curPay.Err(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	trxIDs := make([]string, 0, len(trxIDSet))
	for id := range trxIDSet {
		trxIDs = append(trxIDs, id)
	}

	trxMap := map[string]models.Transaksi{}
	if len(trxIDs) > 0 {
		curTrx, err := trxColl.Find(ctx, bson.M{"_id": bson.M{"$in": trxIDs}})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		defer curTrx.Close(ctx)
		for curTrx.Next(ctx) {
			var t models.Transaksi
			if err := curTrx.Decode(&t); err == nil {
				trxMap[t.ID] = t
			}
		}
	}

	shipByTrx := map[string]models.Pengiriman{}
	userIDSet := map[string]struct{}{}
	pelangganIDSet := map[string]struct{}{}
	produkIDSet := map[string]struct{}{}

	// collect pelanggan/kasir/product IDs from transaksi
	for _, t := range trxMap {
		if t.PelangganID != "" {
			pelangganIDSet[t.PelangganID] = struct{}{}
		}
		if t.KasirID != "" {
			userIDSet[t.KasirID] = struct{}{}
		}
		for _, it := range t.Items {
			if it.NamaProduk == "" && it.ProdukID != "" {
				produkIDSet[it.ProdukID] = struct{}{}
			}
		}
	}

	// fetch pengiriman in bulk
	if len(trxIDs) > 0 {
		curShip, err := pengirimanColl.Find(ctx, bson.M{"transaksi_id": bson.M{"$in": trxIDs}})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		defer curShip.Close(ctx)
		for curShip.Next(ctx) {
			var s models.Pengiriman
			if err := curShip.Decode(&s); err == nil {
				// keep first record per transaksi_id (match UI behavior)
				if _, exists := shipByTrx[s.TransaksiID]; !exists {
					shipByTrx[s.TransaksiID] = s
				}
				if s.DriverID != "" {
					userIDSet[s.DriverID] = struct{}{}
				}
			}
		}
	}

	// fetch pelanggan in bulk
	pelangganIDs := make([]string, 0, len(pelangganIDSet))
	for id := range pelangganIDSet {
		pelangganIDs = append(pelangganIDs, id)
	}
	pelangganMap := map[string]models.Pelanggan{}
	if len(pelangganIDs) > 0 {
		curPel, err := pelangganColl.Find(ctx, bson.M{"_id": bson.M{"$in": pelangganIDs}})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		defer curPel.Close(ctx)
		for curPel.Next(ctx) {
			var p models.Pelanggan
			if err := curPel.Decode(&p); err == nil {
				pelangganMap[p.ID] = p
			}
		}
	}

	// fetch users (kasir + driver) in bulk
	userIDs := make([]string, 0, len(userIDSet))
	for id := range userIDSet {
		userIDs = append(userIDs, id)
	}
	userMap := map[string]models.User{}
	if len(userIDs) > 0 {
		curUser, err := userColl.Find(ctx, bson.M{"_id": bson.M{"$in": userIDs}})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		defer curUser.Close(ctx)
		for curUser.Next(ctx) {
			var u models.User
			if err := curUser.Decode(&u); err == nil {
				userMap[u.ID] = u
			}
		}
	}

	// fetch produk (fallback name when transaksi item doesn't include NamaProduk)
	produkIDs := make([]string, 0, len(produkIDSet))
	for id := range produkIDSet {
		produkIDs = append(produkIDs, id)
	}
	produkNameMap := map[string]string{}
	if len(produkIDs) > 0 {
		curProduk, err := produkColl.Find(ctx, bson.M{"_id": bson.M{"$in": produkIDs}})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		defer curProduk.Close(ctx)
		for curProduk.Next(ctx) {
			var pr models.Produk
			if err := curProduk.Decode(&pr); err == nil {
				produkNameMap[pr.ID] = pr.NamaProduk
			}
		}
	}

	// ===============================
	// BUILD EXCEL
	// ===============================
	f := excelize.NewFile()
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})

	writeHeaders := func(sheet string, headers []string) {
		for i, h := range headers {
			cell, _ := excelize.CoordinatesToCellName(i+1, 1)
			f.SetCellValue(sheet, cell, h)
			f.SetCellStyle(sheet, cell, cell, headerStyle)
		}
	}

	sheetDetail := "Detail Produk"
	f.SetSheetName("Sheet1", sheetDetail)
	writeHeaders(sheetDetail, []string{
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
		"Status Pembayaran",
	})

	sheetRingkas := "Ringkasan Transaksi"
	f.NewSheet(sheetRingkas)
	writeHeaders(sheetRingkas, []string{
		"ID Transaksi",
		"Tanggal",
		"Pelanggan",
		"Total Produk",
		"Ongkir",
		"Total Toko",
		"Status Pembayaran",
	})

	// Sheet 2: Ringkasan Transaksi (loop pembayaran)
	rowR := 2
	for _, pay := range payments {
		tx, hasTx := trxMap[pay.TransaksiID]
		ship := shipByTrx[pay.TransaksiID]
		ongkir := ship.Ongkir
		totalToko := pay.TotalBayar - ongkir
		if totalToko < 0 {
			totalToko = 0
		}

		pelName := ""
		kasirName := ""
		if hasTx {
			if p, ok := pelangganMap[tx.PelangganID]; ok {
				pelName = p.Nama
			} else {
				pelName = tx.PelangganID
			}
			if u, ok := userMap[tx.KasirID]; ok {
				kasirName = u.Nama
			} else {
				kasirName = tx.KasirID
			}
		}
		_ = kasirName // kept for parity with detail sheet; not used in ringkasan columns

		values := []interface{}{
			pay.ID,
			pay.CreatedAt.Format("02-01-2006"),
			pelName,
			func() int {
				if hasTx {
					return tx.TotalProduk
				}
				return 0
			}(),
			ongkir,
			totalToko,
			pay.Status,
		}
		for i, v := range values {
			cell, _ := excelize.CoordinatesToCellName(i+1, rowR)
			f.SetCellValue(sheetRingkas, cell, v)
		}
		rowR++
	}

	// Sheet 1: Detail Produk (loop pembayaran -> transaksi -> items)
	rowD := 2
	totalSubtotal := 0.0
	totalOngkir := 0.0
	seenOngkirByTrx := map[string]struct{}{}
	for _, pay := range payments {
		tx, hasTx := trxMap[pay.TransaksiID]
		if !hasTx || len(tx.Items) == 0 {
			continue
		}
		ship := shipByTrx[pay.TransaksiID]
		ongkir := ship.Ongkir

		pelName := ""
		if p, ok := pelangganMap[tx.PelangganID]; ok {
			pelName = p.Nama
		} else {
			pelName = tx.PelangganID
		}
		kasirName := ""
		if u, ok := userMap[tx.KasirID]; ok {
			kasirName = u.Nama
		} else {
			kasirName = tx.KasirID
		}
		driverName := ""
		if ship.DriverID != "" {
			if u, ok := userMap[ship.DriverID]; ok {
				driverName = u.Nama
			} else {
				driverName = ship.DriverID
			}
		}

		for idx, it := range tx.Items {
			namaProduk := it.NamaProduk
			if namaProduk == "" {
				if n, ok := produkNameMap[it.ProdukID]; ok && n != "" {
					namaProduk = n
				} else {
					namaProduk = it.ProdukID
				}
			}
			subtotal := float64(it.Jumlah) * it.Harga
			totalSubtotal += subtotal
			if idx == 0 {
				if _, already := seenOngkirByTrx[pay.TransaksiID]; !already {
					totalOngkir += ongkir
					seenOngkirByTrx[pay.TransaksiID] = struct{}{}
				}
			}
			ongkirCell := func() interface{} {
				if idx == 0 {
					return ongkir
				}
				return ""
			}()

			values := []interface{}{
				pay.ID,
				pay.CreatedAt.Format("02-01-2006 15:04"),
				pelName,
				kasirName,
				driverName,
				ship.Jenis,
				namaProduk,
				it.Jumlah,
				it.Harga,
				subtotal,
				ongkirCell,
				pay.Status,
			}
			for i, v := range values {
				cell, _ := excelize.CoordinatesToCellName(i+1, rowD)
				f.SetCellValue(sheetDetail, cell, v)
			}
			rowD++
		}
	}

	// Tambahkan ringkasan total di bawah tabel (sesuai permintaan)
	// - Total Subtotal: jumlah semua subtotal item
	// - Total Ongkir: dijumlah sekali per transaksi
	// - Total Bayar: subtotal + ongkir
	summaryRow := rowD + 1
	f.SetCellValue(sheetDetail, fmt.Sprintf("I%d", summaryRow), "TOTAL SUBTOTAL")
	f.SetCellValue(sheetDetail, fmt.Sprintf("J%d", summaryRow), totalSubtotal)
	summaryRow++
	f.SetCellValue(sheetDetail, fmt.Sprintf("I%d", summaryRow), "TOTAL ONGKIR")
	f.SetCellValue(sheetDetail, fmt.Sprintf("J%d", summaryRow), totalOngkir)
	summaryRow++
	f.SetCellValue(sheetDetail, fmt.Sprintf("I%d", summaryRow), "TOTAL BAYAR (SUBTOTAL+ONGKIR)")
	f.SetCellValue(sheetDetail, fmt.Sprintf("J%d", summaryRow), totalSubtotal+totalOngkir)

	f.AutoFilter(sheetDetail, "A1:L1", []excelize.AutoFilterOptions{})
	f.SetPanes(sheetDetail, &excelize.Panes{Freeze: true, Split: true, YSplit: 1})

	f.AutoFilter(sheetRingkas, "A1:G1", []excelize.AutoFilterOptions{})
	f.SetPanes(sheetRingkas, &excelize.Panes{Freeze: true, Split: true, YSplit: 1})

	// Response
	f.SetActiveSheet(0)
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", "attachment; filename=laporan_mbg.xlsx")
	buf, _ := f.WriteToBuffer()
	return c.Send(buf.Bytes())
}
