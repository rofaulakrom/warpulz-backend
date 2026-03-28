package models

import (
	"time"
)

// Product adalah representasi tabel 'products' di database
type Product struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(255);not null" json:"name"`
	SKU       string    `gorm:"type:varchar(100);unique;not null" json:"sku"` // Kode barang unik
	Price     int64     `gorm:"not null" json:"price"`                        // Harga dalam Rupiah penuh
	Stock     int       `gorm:"not null;default:0" json:"stock"`              // Jumlah stok
	Category  string    `gorm:"type:varchar(100)" json:"category"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}