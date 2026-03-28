package models

import (
	"time"
)

type Expense struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Description string    `json:"description"` // Contoh: "Beli Telur 2 Kg", "Bayar Listrik"
	Category    string    `json:"category"`    // Contoh: "Bahan Baku", "Operasional", "Lain-lain"
	Amount      int       `json:"amount"`      // Nominal pengeluaran
	ExpenseDate string    `json:"expense_date"`// Tanggal pengeluaran (YYYY-MM-DD)
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}