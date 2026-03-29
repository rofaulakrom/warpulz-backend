package main

import (
	"net/http"
	"os" // Untuk membaca port dari sistem (HuggingFace/Railway)

	// SESUAIKAN IMPORT INI dengan path project Anda
	"github.com/rofaulakrom/warpulz-backend/database"
	"github.com/rofaulakrom/warpulz-backend/handlers"
	"github.com/rofaulakrom/warpulz-backend/models"

	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// 1. Inisialisasi Database
	database.ConnectDB()

	// Otomatis membuat tabel expenses di database
	database.DB.AutoMigrate(&models.Expense{})

	e := echo.New()

	// 2. Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// 3. CORS (Bebas diakses dari mana saja termasuk Vercel nanti)
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
		},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
		},
	}))

	// Route Test (Health Check untuk Hugging Face)
	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "Halo! Backend WARPULZ Ready 🚀",
		})
	})

	// === ROUTE API ===
	e.POST("/login", handlers.LoginAdmin)

	// Endpoint Publik
	e.GET("/products", handlers.GetAllProducts)
	e.POST("/orders", handlers.CreateOrder)
	e.POST("/payment-notification", handlers.HandleMidtransNotification)
	e.GET("/orders/:id/receipt", handlers.DownloadReceipt)

	// === KUNCI KEAMANAN JWT ===
	jwtMiddleware := echojwt.WithConfig(echojwt.Config{
		SigningKey: []byte("WARPULZ_SUPER_SECRET_KEY_2026"),
	})

	// Endpoint yang Dilindungi
	e.POST("/products", handlers.CreateProduct, jwtMiddleware)
	e.PUT("/products/:id", handlers.UpdateProduct, jwtMiddleware)
	e.DELETE("/products/:id", handlers.DeleteProduct, jwtMiddleware)

	e.GET("/orders", handlers.GetAllOrders, jwtMiddleware)
	e.PUT("/orders/:id/status", handlers.UpdateOrderStatus, jwtMiddleware)

	e.POST("/expenses", handlers.CreateExpense, jwtMiddleware)
	e.GET("/expenses", handlers.GetAllExpenses, jwtMiddleware)
	e.PUT("/expenses/:id", handlers.UpdateExpense, jwtMiddleware)
	e.DELETE("/expenses/:id", handlers.DeleteExpense, jwtMiddleware)

	// === PERBAIKAN UNTUK HUGGING FACE ===
	// Ambil port dari Environment, jika kosong pakai 8080 (lokal)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// WAJIB: Gunakan 0.0.0.0 agar Hugging Face bisa mendeteksi servernya di IPv4
	e.Logger.Fatal(e.Start("0.0.0.0:" + port))
}