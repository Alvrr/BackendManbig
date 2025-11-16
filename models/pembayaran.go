package models

type ItemProduk struct {
	IDProduk   string `json:"id_produk" bson:"id_produk"`
	NamaProduk string `json:"nama_produk" bson:"nama_produk"`
	Harga      int    `json:"harga" bson:"harga"`
	Jumlah     int    `json:"jumlah" bson:"jumlah"`
	Subtotal   int    `json:"subtotal" bson:"subtotal"`
}

type Pembayaran struct {
	ID              string       `json:"id" bson:"_id"`
	IDPelanggan     string       `json:"id_pelanggan" bson:"id_pelanggan" validate:"required"`
	IDKasir         string       `json:"id_kasir" bson:"id_kasir"`
	NamaKasir       string       `json:"nama_kasir" bson:"nama_kasir"`
	IDDriver        string       `json:"id_driver,omitempty" bson:"id_driver,omitempty"`
	NamaDriver      string       `json:"nama_driver,omitempty" bson:"nama_driver,omitempty"`
	JenisPengiriman string       `json:"jenis_pengiriman" bson:"jenis_pengiriman"`
	Produk          []ItemProduk `json:"produk" bson:"produk" validate:"required"`
	TotalBayar      int          `json:"total_bayar" bson:"total_bayar"`
	Ongkir          int          `json:"ongkir,omitempty" bson:"ongkir,omitempty"`
	Tanggal         string       `json:"tanggal,omitempty" bson:"tanggal"`
	Status          string       `json:"status" bson:"status"`
}
