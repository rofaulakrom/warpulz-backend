package handlers

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// Gunakan kata sandi acak yang panjang untuk produksi. Simpan di .env!
const jwtSecretKey = "WARPULZ_SUPER_SECRET_KEY_2026"

func LoginAdmin(c echo.Context) error {
	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.Bind(&credentials); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Data tidak valid"})
	}

	// Simulasi pengecekan ke database (Untuk MVP, kita hardcode kredensial Admin)
	if credentials.Username != "admin" || credentials.Password != "warpulz123" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Username atau Password salah!"})
	}

	// Jika sukses, buat struktur Token JWT (berlaku 24 jam)
	claims := jwt.MapClaims{
		"name": "Admin WARPULZ",
		"role": "admin",
		"exp":  time.Now().Add(time.Hour * 24).Unix(), // Waktu kadaluarsa
	}

	// Proses penandatanganan Token menggunakan Kunci Rahasia
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(jwtSecretKey))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal membuat token"})
	}

	// Kirim token ke Frontend
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Login berhasil",
		"token":   t,
	})
}