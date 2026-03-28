package handlers

import (
	"net/http"
	"strconv"

	"github.com/rofaulakrom/warpulz-backend/database"
	"github.com/rofaulakrom/warpulz-backend/models"

	"github.com/labstack/echo/v4"
)

// 1. Create Expense
func CreateExpense(c echo.Context) error {
	expense := new(models.Expense)
	if err := c.Bind(expense); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Data tidak valid"})
	}

	if err := database.DB.Create(&expense).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menyimpan pengeluaran"})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "✅ Pengeluaran berhasil dicatat",
		"data":    expense,
	})
}

// 2. Get All Expenses
func GetAllExpenses(c echo.Context) error {
	var expenses []models.Expense
	// Mengurutkan dari yang terbaru
	if err := database.DB.Order("expense_date desc, created_at desc").Find(&expenses).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal mengambil data pengeluaran"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": expenses,
	})
}

// 3. Update Expense
func UpdateExpense(c echo.Context) error {
	id := c.Param("id")
	expenseId, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "ID tidak valid"})
	}

	var expense models.Expense
	if err := database.DB.First(&expense, expenseId).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "Data pengeluaran tidak ditemukan"})
	}

	if err := c.Bind(&expense); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Data tidak valid"})
	}

	database.DB.Save(&expense)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "✅ Pengeluaran berhasil diperbarui",
		"data":    expense,
	})
}

// 4. Delete Expense
func DeleteExpense(c echo.Context) error {
	id := c.Param("id")
	expenseId, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "ID tidak valid"})
	}

	if err := database.DB.Delete(&models.Expense{}, expenseId).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal menghapus data"})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "✅ Data pengeluaran berhasil dihapus",
	})
}