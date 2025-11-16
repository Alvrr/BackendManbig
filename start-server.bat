@echo off
echo 🚀 Memulai Backend MBG Server...
echo.
echo Server akan berjalan di: http://localhost:5000
echo Swagger UI tersedia di: http://localhost:5000/swagger/index.html
echo.
echo ⚡ PENTING: Setelah server berjalan, buka browser dan:
echo 1. Login di /auth/login dengan kredensial berikut:
echo    - Admin: admin@email.com / password123
echo    - Kasir: kasir@email.com / password123
echo    - Driver: driver@email.com / password123
echo 2. Copy token dari response login
echo 3. Klik tombol "Authorize" di Swagger UI
echo 4. Masukkan: Bearer {your-token}
echo.
echo Press Ctrl+C untuk menghentikan server
echo.

go run main.go
