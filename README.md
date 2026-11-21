# laundrypos-be

Backend POS laundry: Go 1.22 + MongoDB (Atlas), di-deploy ke Vercel sebagai serverless function.

## Menjalankan lokal

1. Salin `.env.example` jadi `.env`, lalu isi:
   - `MONGO_URI`: string koneksi MongoDB Atlas
   - `JWT_SECRET_KEY`: secret acak minimal 32 karakter (`openssl rand -hex 32`)
   - `MIDTRANS_SERVER_KEY` / `MIDTRANS_CLIENT_KEY`: dari dashboard Midtrans
   - `CORS_ALLOWED_ORIGINS`: daftar origin frontend, dipisah koma
2. Jalankan server: **`vercel dev --listen 8082`** (butuh `npm i -g vercel` dan sudah `vercel link`).
   Fungsi di `api/` memakai `package handler` sesuai runtime Vercel, jadi `go run` tidak dipakai lagi.
3. Buat user admin: `ADMIN_USERNAME=admin ADMIN_PASSWORD=xxxxxxxx go run ./cmd/seed`
4. Data lama (opsional): `go run ./cmd/migrate` memigrasi status lama dan mengisi nomor nota retroaktif.

## Test

```bash
go test ./...                                    # unit test murni
TEST_MONGO_URI=mongodb+srv://... go test ./...   # + test integrasi (database laundry-pos-test)
go vet ./...
```

## Ringkasan API

Semua respons berbentuk `{success, message?, data?, meta?}`. Autentikasi JWT via header `Authorization: Bearer <token>`.

| Method & path | Akses | Fungsi |
|---|---|---|
| POST /register, /login, /logout | publik / token | registrasi (role staff), login, mencabut token |
| GET /health | publik | cek server |
| GET/POST /customers, GET/PUT/DELETE /customers/{id} | token | CRUD pelanggan (list paginated, filter `q`) |
| GET /services, GET /services/{id} | token | daftar layanan |
| POST/PUT/DELETE /services... | admin | kelola layanan |
| GET/POST /transactions, GET/PUT /transactions/{id} | token | transaksi; edit item hanya saat `pending` |
| PATCH /transactions/{id}/status | token | siklus status: pending → diproses → selesai → diambil (+ batal) |
| DELETE /transactions/{id} | admin | hapus transaksi |
| POST /payments/cash | token | bayar tunai + kembalian, anti dobel bayar |
| POST /payments/midtrans | token | Snap Midtrans (pakai ulang tagihan pending) |
| GET /payments, GET /payments/{orderID} | token | riwayat pembayaran |
| POST /webhook/midtrans | publik + signature SHA512 | update status pembayaran & transaksi |
| GET /reports/summary | admin | pendapatan hari/bulan, pesanan hari ini, belum lunas |

## Deploy (Vercel)

`vercel.json` mem-build `api/main.go` (runtime Go menuntut `package handler`, bukan `package main`). Set semua env di dashboard Vercel (bukan di file), termasuk `APP_TZ=Asia/Jakarta` supaya batas "hari ini" di laporan mengikuti waktu toko, dan `MIDTRANS_ENV=production` saat sudah go-live. Hindari var `DB_NAME` kosong/salah: nilai itu menentukan nama database Mongo yang dipakai (default `laundry-pos`). **Penting:** kredensial Atlas dan kunci Midtrans lama pernah ter-commit ke git history, ganti/rotasi keduanya sebelum produksi.

Catatan: pembayaran cash dan webhook memakai transaksi Mongo (`WithTransaction`) yang butuh replica set. MongoDB Atlas sudah replica set; Mongo lokal standalone tidak akan bisa menjalankan pembayaran cash.

Catatan keamanan: `/register` publik sehingga siapa pun dengan URL bisa mendaftar sebagai staf. Kalau app dipakai di internet, pertimbangkan menambah kode undangan atau menonaktifkan pendaftaran mandiri.
