# 💸 PayAja - High Performance Wallet Service

**PayAja** adalah backend service untuk dompet digital (E-Wallet) yang dibangun menggunakan **Go (Golang)**. Proyek ini dirancang khusus untuk menangani transaksi keuangan dengan **konkurensi tinggi**, menjamin integritas data (anti-race condition), dan mengikuti standar arsitektur industri modern.

---

## 🚀 Fitur Unggulan

-   **Clean Architecture (DDD):** Pemisahan kode yang tegas antara *Handler, Usecase, Repository,* dan *Domain*. Kode modular, mudah dites, dan *maintainable*.
-   **Concurrency Safe (Anti-Race Condition):** Menggunakan mekanisme **Pessimistic Locking** (`SELECT ... FOR UPDATE`) di level database untuk mencegah saldo hilang saat ribuan request masuk bersamaan.
-   **High Performance gRPC:** Komunikasi antar service (Gateway ke Wallet Service) menggunakan protokol gRPC yang cepat dan ringan.
-   **REST API Gateway:** Menjembatani request HTTP/JSON dari client ke sistem internal gRPC.
-   **Swagger Documentation:** Dokumentasi API interaktif yang tergenerate otomatis.
-   **Connection Pooling:** Manajemen koneksi database yang efisien untuk menangani *backpressure*.
-   **Docker Ready:** Siap dideploy di mana saja menggunakan Docker Compose.

---

## 🛠️ Tech Stack

| Komponen | Teknologi | Deskripsi |
| :--- | :--- | :--- |
| **Language** | Go (1.24) | Backend Logic |
| **Framework** | Gin Gonic | API Gateway & HTTP Routing |
| **RPC** | gRPC & Protobuf | Internal Service Communication |
| **Database** | PostgreSQL 15 | Primary Storage |
| **ORM** | GORM | Database Access & Transaction Mgmt |
| **Docs** | Swagger (Swag) | API Documentation |
| **Infra** | Docker | Containerization |

---

## 🏗️ Struktur Proyek (Clean Architecture)

```text
pay-aja/
├── api-gateway/            # HTTP Entry point (Gin Framework)
│   ├── docs/               # Generated Swagger docs
│   └── main.go             # Gateway Logic & Swagger setup
├── client/                 # Tools untuk pengujian manual
│   └── stress.go           # Script Stress Test (Race Condition Proof)
├── proto/                  # Definisi kontrak gRPC (.proto)
├── wallet-service/         # Core Microservice
│   ├── internal/
│   │   ├── domain/         # Entity & Interface (Kontrak Bisnis Murni)
│   │   ├── repository/     # Implementasi DB (Postgres + Locking)
│   │   ├── usecase/        # Logic Bisnis (Validasi, Flow)
│   │   └── handler/        # gRPC Server Handler
│   └── main.go             # Dependency Injection & Wiring
└── docker-compose.yml      # Orchestration
```
---

## 🏃‍♂️ Cara Menjalankan

### Prasyarat

* Docker & Docker Compose terinstall di laptop.

### 1. Jalankan Aplikasi

Cukup jalankan satu perintah ini. Docker akan membuild service backend, gateway, dan database secara otomatis.

```bash
docker-compose up --build

```

*Tunggu hingga muncul log: `🔥 API Gateway running on :8080` dan `🚀 Wallet Service (Clean DDD) running on :50051`.*

### 2. Akses Dokumentasi API (Swagger)

Buka browser dan kunjungi:
👉 **[http://localhost:8080/swagger/index.html](https://www.google.com/search?q=http://localhost:8080/swagger/index.html)**

Anda bisa mencoba endpoint berikut langsung dari browser:

* `GET /balance/{user_id}` : Cek saldo (Auto-register user baru).
* `POST /topup` : Isi saldo.
* `POST /transfer` : Kirim uang antar user.

---

## 🛡️ Pengujian Keamanan (Stress Test)

Proyek ini dilengkapi dengan alat uji beban untuk membuktikan bahwa sistem **TIDAK** memiliki bug *Race Condition*.

**Skenario Tes:**

* **1.000 Request** TopUp dikirim secara bersamaan (Concurrent Goroutines).
* Setiap request bernilai **Rp 1.000**.
* **Target:** Saldo akhir harus bertambah tepat **Rp 1.000.000**.

**Cara Menjalankan Tes:**
Buka terminal baru (pastikan server docker sedang berjalan), lalu ketik:

```bash
go run client/stress.go

```

**Hasil Output:**

```text
🚀 MEMULAI Test: 1000 Request x Rp 1000...
...
Total Masuk: Rp 1000000
SISTEM AMAN! (Tidak ada Race Condition)

```

---

## 🧠 Solusi Teknis: Race Condition

Masalah utama dalam sistem saldo adalah ketika dua transaksi membaca saldo yang sama secara bersamaan sebelum salah satunya menyimpan update.

**Solusi di PayAja:**
Saya menggunakan **Pessimistic Locking** pada repository layer.

```go
// wallet-service/internal/repository/postgres_gorm.go

func (r *walletRepository) UpdateBalance(...) error {
    return r.db.Transaction(func(tx *gorm.DB) error {
        // 🔒 "SELECT ... FOR UPDATE"
        // Database akan mengunci baris user ini sampai transaksi selesai.
        // Request lain harus mengantre (queue), mencegah data tertimpa.
        if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
            Where("user_id = ?", userID).
            First(&wallet).Error; err != nil {
            return err
        }
        
        wallet.Balance += amount
        return tx.Save(&wallet).Error
    })
}

```



