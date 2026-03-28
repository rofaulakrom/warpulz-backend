package handlers

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/rofaulakrom/warpulz-backend/database"
	"github.com/rofaulakrom/warpulz-backend/models"
	"gorm.io/gorm"
)

// CreateProduct: Menambah barang baru
func CreateProduct(c echo.Context) error {
	product := new(models.Product)
	if err := c.Bind(product); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "Data tidak valid!",
			"error":   err.Error(),
		})
	}

	result := database.DB.Create(&product)
	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Gagal menyimpan ke database",
			"error":   result.Error.Error(),
		})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "✅ Sukses menambahkan barang!",
		"data":    product,
	})
}

// GetAllProducts: Mengambil semua data barang
func GetAllProducts(c echo.Context) error {
	var products []models.Product
	result := database.DB.Find(&products)

	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Gagal mengambil data",
			"error":   result.Error.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "✅ Sukses ambil data barang",
		"data":    products,
	})
}

// UpdateProduct: Mengubah data produk berdasarkan ID
func UpdateProduct(c echo.Context) error {
	id := c.Param("id")
	productId, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "ID Produk tidak valid"})
	}

	var product models.Product
	if err := database.DB.First(&product, productId).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "Produk tidak ditemukan"})
	}

	if err := c.Bind(&product); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Data tidak valid"})
	}

	database.DB.Save(&product)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Produk berhasil diperbarui",
		"data":    product,
	})
}

// DeleteProduct: Menghapus produk
func DeleteProduct(c echo.Context) error {
	id := c.Param("id")
	productId, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "ID Produk tidak valid"})
	}

	result := database.DB.Delete(&models.Product{}, productId)
	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menghapus produk"})
	}

	if result.RowsAffected == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "Produk tidak ditemukan"})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Produk berhasil dihapus",
	})
}

// ReduceStock: Fungsi ini dipanggil saat order LUNAS (Paid)
// Menggunakan huruf kapital 'R' agar bisa dipanggil dari order_handler.go
func ReduceStock(productDetails string) {
	// Membagi string berdasarkan pemisah "|"
	items := strings.Split(productDetails, " | ")

	for _, item := range items {
		// Regex untuk menangkap: Nama Produk (nx)
		re := regexp.MustCompile(`^(.*) \((\d+)x\)`)
		match := re.FindStringSubmatch(item)

		if len(match) == 3 {
			productName := strings.TrimSpace(match[1])
			quantity, _ := strconv.Atoi(match[2])

			// Update stok di database secara atomik
			err := database.DB.Model(&models.Product{}).
				Where("name = ?", productName).
				UpdateColumn("stock", gorm.Expr("stock - ?", quantity)).Error
			
			if err != nil {
				fmt.Printf("❌ Gagal potong stok untuk %s: %v\n", productName, err)
			} else {
				fmt.Printf("📉 Stok Berkurang: %s sebanyak %d\n", productName, quantity)
			}
		}
	}
}