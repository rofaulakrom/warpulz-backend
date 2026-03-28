package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Order struct dengan penyesuaian untuk konsep F&B (Cafe/Resto)
type Order struct {
	ID            uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	InvoiceNumber string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"invoice_number"`

	// Data Pelanggan F&B
	CustomerName  string `gorm:"type:varchar(100);not null" json:"customer_name" validate:"required,min=3"`
	CustomerEmail string `gorm:"type:varchar(100);not null" json:"customer_email" validate:"required,email"`
	
	// Data Spesifik F&B
	TableNumber string `gorm:"type:varchar(20);default:'-'" json:"table_number"`
	OrderType   string `gorm:"type:varchar(20);default:'Dine In'" json:"order_type"`

	// Detail Barang 
	ProductName string `gorm:"type:text;not null" json:"product_name"`
	Quantity    int    `gorm:"not null" json:"quantity" validate:"required,min=1"`
	TotalPrice  int64  `gorm:"not null" json:"total_price"` 

	// Status & Audit Log
	Status        string    `gorm:"type:varchar(20);default:'Pending';index" json:"status"`
	PaymentMethod string    `gorm:"type:varchar(50)" json:"payment_method"` 
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (o *Order) BeforeCreate(tx *gorm.DB) (err error) {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return
}