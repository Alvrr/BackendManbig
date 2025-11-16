# 📖 Panduan Menggunakan Swagger UI - Backend MBG

## 🚀 Cara Test API dengan Swagger UI

### Langkah 1: Jalankan Backend Server
```bash
go run main.go
```
Server akan berjalan di: http://localhost:5000

### Langkah 2: Buka Swagger UI
Buka browser dan akses: http://localhost:5000/swagger/index.html

### Langkah 3: Login untuk Mendapatkan Token

1. **Cari endpoint `/auth/login`** di bagian "Authentication"
2. **Klik "Try it out"**
3. **Masukkan kredensial** (gunakan salah satu):
   ```json
   {
     "email": "admin@email.com",
     "password": "password123"
   }
   ```
   atau
   ```json
   {
     "email": "kasir@email.com",
     "password": "password123"
   }
   ```
   atau
   ```json
   {
     "email": "driver@email.com",
     "password": "password123"
   }
   ```
4. **Klik "Execute"**
5. **Copy token** dari response (bagian "token")

### Langkah 4: Set Authorization

1. **Klik tombol "Authorize" ⚡** di pojok kanan atas halaman Swagger
2. **Masukkan:** `Bearer {token-yang-sudah-dicopy}`
   
   **Format yang BENAR:**
   ```
   Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjY3M2ExMjM0NTY3ODkwYWJjZGVmMTIzNCIsIm5hbWEiOiJBZG1pbiBVc2VyIiwiZW1haWwiOiJhZG1pbkBlbWFpbC5jb20iLCJyb2xlIjoiYWRtaW4iLCJleHAiOjE3MDUzMjAwMDB9.signature
   ```
   
   **❌ Format SALAH:**
   ```
   eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...  (tanpa "Bearer")
   bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...  (huruf kecil)
   ```

3. **Klik "Authorize"**
4. **Jika berhasil**, akan muncul ✅ di tombol "Authorize"

### Langkah 5: Test Endpoint Lainnya

Sekarang semua endpoint sudah bisa ditest! Contoh:

- **Data Karyawan** (admin only)
- **CRUD Pelanggan** 
- **CRUD Produk**
- **Transaksi Pembayaran**
- **Riwayat & Laporan**

## 🔧 Troubleshooting

### Masalah: "Token tidak ditemukan atau format salah"

**Solusi:**
1. Pastikan sudah login di `/auth/login` dan mendapat response token
2. Copy token **lengkap** dari response
3. Di Authorization, masukkan: `Bearer {token}` (ada spasi setelah Bearer)
4. Pastikan tidak ada karakter tambahan atau spasi di awal/akhir

### Masalah: "401 Unauthorized"

**Solusi:**
1. Token mungkin expired, login ulang
2. Cek format Authorization header
3. Pastikan role user sesuai dengan endpoint yang diakses

### Masalah: "403 Forbidden" 

**Solusi:**
1. Cek role permissions:
   - **Admin**: Akses semua endpoint
   - **Kasir**: CRUD produk, pelanggan, pembayaran, riwayat
   - **Driver**: Hanya lihat & complete pembayaran assigned

## 📋 Contoh Input untuk Test

### Create Pelanggan:
```json
{
  "nama": "John Doe",
  "email": "john.doe@email.com",
  "no_hp": "081234567890",
  "alamat": "Jl. Sudirman No. 123, Jakarta Pusat"
}
```

### Create Produk:
```json
{
  "nama_produk": "Sabun Cuci Super Clean",
  "kategori": "Deterjen",
  "harga": 15000,
  "stok": 50,
  "deskripsi": "Sabun cuci baju berkualitas tinggi"
}
```

### Create Karyawan (Admin only):
```json
{
  "nama": "Budi Santoso",
  "email": "kasir01@email.com",
  "password": "password123",
  "role": "kasir",
  "no_hp": "081234567890",
  "alamat": "Jl. Merdeka No. 45, Bandung",
  "status": "aktif"
}
```

---

🎉 **Selamat! Sekarang Anda bisa test semua API dengan mudah menggunakan Swagger UI!**
