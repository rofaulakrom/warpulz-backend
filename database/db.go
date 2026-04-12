package database

import (
	"fmt"
	"log"
	"os"

	// GANTI INI SESUAI NAMA MODULE ANDA DI go.mod
	"github.com/rofaulakrom/warpulz-backend/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	// 1. PERBAIKAN: Jangan matikan server jika tidak ada file .env (karena di Railway memang tidak ada file .env)
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Info: File .env tidak ditemukan, menggunakan Environment Variables dari sistem Cloud (Railway).")
	}

	var dsn string

	// 2. PERBAIKAN: Baca dari DATABASE_URL (Format dari Supabase & Railway)
	dbUrl := os.Getenv("DATABASE_URL")

	if dbUrl != "" {
		// Jika jalan di Railway, otomatis pakai URL Supabase
		dsn = dbUrl
	} else {
		// Jika jalan di Laptop (Lokal), pakai variabel yang dipisah
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"),
			os.Getenv("DB_PORT"),
		)
	}

	connection, err := gorm.Open(postgres.New(postgres.Config{
        DSN:                  dsn,
        PreferSimpleProtocol: true, // <--- Mencegah error antrean di Supabase (42P05)
    }), &gorm.Config{})
    
    if err != nil {
        log.Fatal("❌ Gagal konek ke database: ", err)
    }

    DB = connection
    fmt.Println("✅ Sukses terhubung ke Database PostgreSQL!")
	
	// === AUTO MIGRATION ===
	fmt.Println("⏳ Sedang melakukan migrasi database...")
	// Saya gabungkan semua tabel di sini agar lebih rapi
	err = DB.AutoMigrate(&models.Product{}, &models.Order{}, &models.Expense{})
	if err != nil {
		fmt.Println("❌ Gagal migrasi:", err)
	} else {
		fmt.Println("✅ Sukses migrasi tabel!")
	}
}