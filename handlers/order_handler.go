package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rofaulakrom/warpulz-backend/database"
	"github.com/rofaulakrom/warpulz-backend/models"

	"github.com/go-playground/validator/v10"
	"github.com/jung-kurt/gofpdf"
	"github.com/labstack/echo/v4"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
	"gopkg.in/gomail.v2"
)

var validate = validator.New()

var (
	MidtransServerKey = os.Getenv("MIDTRANS_SERVER_KEY")
	// Variabel Email dihapus dari sini agar dipanggil langsung di dalam fungsinya
)

// --- CONFIG CUSTOM DATA TOKO ---
// Anda bisa mengubah isi data ini sesuai dengan toko Anda
var StoreConfig = map[string]string{
	"Name":      "Warkop Pulang (WARPULZ)",
	"Address":   "Jl. Sapta Marga, Campaka, Kec. Andir, Kota Bandung, Jawa Barat",
	"Phone":     "WA: 0895-3643-70623",
	"WiFi_User": "WARPULZ_GUEST",
	"WiFi_Pass": "nyemekjosjis",
	"Greeting":  "Jika nyaman beritahu teman, jika tidak beritahu kami. \nTerima kasih! Sampai jumpa kembali.",
}

// ==========================================
// FUNGSI BARU: GENERATE STRUK PDF
// ==========================================
func generateReceiptPDF(order *models.Order) (string, error) {
	pdf := gofpdf.New("P", "mm", "A5", "")
	pdf.AddPage()

	// Header Toko
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(0, 8, StoreConfig["Name"], "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(0, 5, StoreConfig["Address"], "", 1, "C", false, 0, "")
	pdf.CellFormat(0, 5, StoreConfig["Phone"], "", 1, "C", false, 0, "")
	pdf.Ln(5)

	// Garis Pemisah
	pdf.SetLineWidth(0.5)
	pdf.Line(10, pdf.GetY(), 138, pdf.GetY())
	pdf.Ln(3)

	// Detail Order
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(0, 6, fmt.Sprintf("INVOICE: %s", order.InvoiceNumber), "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(0, 5, fmt.Sprintf("Tanggal : %s", order.CreatedAt.Format("02 Jan 2006 15:04")), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 5, fmt.Sprintf("Kasir   : System (%s)", order.PaymentMethod), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 5, fmt.Sprintf("Tipe    : %s (Meja: %s)", order.OrderType, order.TableNumber), "", 1, "L", false, 0, "")
	pdf.Ln(3)
	pdf.Line(10, pdf.GetY(), 138, pdf.GetY())
	pdf.Ln(3)

	// Daftar Item
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(0, 6, "DAFTAR PESANAN:", "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 10)

	items := strings.Split(order.ProductName, " | ")
	for _, item := range items {
		pdf.MultiCell(0, 5, "- "+item, "", "L", false)
	}

	pdf.Ln(3)
	pdf.Line(10, pdf.GetY(), 138, pdf.GetY())
	pdf.Ln(3)

	// Total Harga
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(80, 8, "TOTAL BAYAR", "", 0, "L", false, 0, "")
	pdf.CellFormat(48, 8, fmt.Sprintf("Rp %d", order.TotalPrice), "", 1, "R", false, 0, "")

	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(80, 6, "STATUS", "", 0, "L", false, 0, "")
	pdf.CellFormat(48, 6, strings.ToUpper(order.Status), "", 1, "R", false, 0, "")

	pdf.Ln(5)
	pdf.Line(10, pdf.GetY(), 138, pdf.GetY())
	pdf.Ln(5)

	// Footer & WiFi
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(0, 6, "INFO WIFI", "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(0, 5, fmt.Sprintf("SSID: %s | Pass: %s", StoreConfig["WiFi_User"], StoreConfig["WiFi_Pass"]), "", 1, "C", false, 0, "")

	pdf.Ln(5)
	pdf.SetFont("Arial", "I", 10)
	pdf.CellFormat(0, 6, StoreConfig["Greeting"], "", 1, "C", false, 0, "")

	fileName := fmt.Sprintf("Struk_%s.pdf", order.InvoiceNumber)
	err := pdf.OutputFileAndClose(fileName)
	return fileName, err
}

// ==========================================
// FUNGSI BARU: ENDPOINT DOWNLOAD PDF
// ==========================================
func DownloadReceipt(c echo.Context) error {
	id := c.Param("id")
	var order models.Order
	if err := database.DB.Where("id = ?", id).First(&order).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "Order tidak ditemukan"})
	}

	filePath, err := generateReceiptPDF(&order)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal membuat struk PDF"})
	}

	return c.Attachment(filePath, filePath)
}

// ==========================================
// FUNGSI BARU: KIRIM EMAIL DENGAN LAMPIRAN PDF
// ==========================================
func sendReceiptEmail(order *models.Order) {
	pdfPath, err := generateReceiptPDF(order)
	if err != nil {
		fmt.Println("❌ Gagal membuat PDF:", err)
		return
	}

	// PERBAIKAN: Ambil variabel environment tepat saat fungsi dijalankan
	EmailSender := os.Getenv("EMAIL_SENDER")
	EmailPassword := os.Getenv("EMAIL_PASSWORD")

	m := gomail.NewMessage()
	m.SetHeader("From", EmailSender)
	m.SetHeader("To", order.CustomerEmail)
	m.SetHeader("Subject", "✅ Pembayaran Berhasil - "+order.InvoiceNumber)

	htmlBody := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; text-align: center;">
			<h2 style="color: #16a34a;">Pembayaran Berhasil!</h2>
			<p>Halo, <b>%s</b>!</p>
			<p>Terima kasih, pembayaran Anda untuk pesanan <b>%s</b> telah kami terima.</p>
			<p>Silakan lihat lampiran PDF pada email ini untuk mengunduh Struk Digital Anda (termasuk password WiFi).</p>
			<br/>
			<p style="color: #64748b; font-size: 12px;">%s</p>
		</div>
	`, order.CustomerName, order.InvoiceNumber, StoreConfig["Greeting"])

	m.SetBody("text/html", htmlBody)
	m.Attach(pdfPath)

	d := gomail.NewDialer("smtp.gmail.com", 587, EmailSender, EmailPassword)
	if err := d.DialAndSend(m); err != nil {
		fmt.Println("❌ Gagal kirim email lunas:", err)
	} else {
		fmt.Println("✅ Email PDF Struk terkirim ke:", order.CustomerEmail)
	}

	// Hapus PDF lokal setelah terkirim
	os.Remove(pdfPath)
}

// Function untuk Kirim Email
func sendEmailNotification(order *models.Order, paymentURL string) {
	// PERBAIKAN: Ambil variabel environment tepat saat fungsi dijalankan
	EmailSender := os.Getenv("EMAIL_SENDER")
	EmailPassword := os.Getenv("EMAIL_PASSWORD")

	m := gomail.NewMessage()
	m.SetHeader("From", EmailSender)
	m.SetHeader("To", order.CustomerEmail)
	m.SetHeader("Subject", "Struk Pesanan - "+order.InvoiceNumber)

	paymentActionHTML := ""
	if order.PaymentMethod == "Cash" {
		paymentActionHTML = `<div style="background-color: #fef3c7; border-left: 4px solid #f59e0b; padding: 15px; margin-top: 20px;">
			<h3 style="color: #d97706; margin: 0 0 5px 0;">Pembayaran Tunai (Di Kasir)</h3>
			<p style="margin: 0; color: #92400e;">Silakan tunjukkan email ini kepada Kasir kami untuk melakukan pembayaran dan memproses pesanan Anda.</p>
		</div>`
	} else {
		paymentActionHTML = fmt.Sprintf(`
			<h3>Silakan selesaikan pembayaran melalui link ini:</h3>
			<a href="%s" style="background-color: #2563eb; color: white; padding: 12px 24px; text-decoration: none; border-radius: 25px; font-weight: bold; display: inline-block;">Bayar Sekarang (QRIS/Transfer)</a>
			<br><br>
			<p style="font-size: 12px; color: #64748b;">Jika tombol tidak berfungsi, copy-paste link ini ke browser: <br> %s</p>
		`, paymentURL, paymentURL)
	}

	htmlBody := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
			<h2 style="color: #2563eb;">Halo, %s!</h2>
			<p>Terima kasih telah memesan di Warpulz.</p>
			<div style="background-color: #f8fafc; padding: 15px; border-radius: 10px; margin-bottom: 20px;">
				<p style="margin: 0 0 10px 0;"><b>Detail Pesanan Anda:</b></p>
				<ul style="list-style-type: none; padding: 0; margin: 0;">
					<li style="margin-bottom: 5px;"><b>Tipe:</b> %s (Meja: %s)</li>
					<li style="margin-bottom: 5px;"><b>Menu:</b> %s</li>
					<li style="margin-bottom: 5px;"><b>Metode Bayar:</b> %s</li>
					<li style="margin-bottom: 5px;"><b>Total:</b> Rp %d</li>
					<li><b>Invoice:</b> %s</li>
				</ul>
			</div>
			%s
		</div>
	`, order.CustomerName, order.OrderType, order.TableNumber, order.ProductName, order.PaymentMethod, order.TotalPrice, order.InvoiceNumber, paymentActionHTML)

	m.SetBody("text/html", htmlBody)
	d := gomail.NewDialer("smtp.gmail.com", 587, EmailSender, EmailPassword)

	if err := d.DialAndSend(m); err != nil {
		fmt.Println("❌ Gagal kirim email:", err)
	} else {
		fmt.Println("✅ Email notifikasi terkirim ke:", order.CustomerEmail)
	}
}

func CreateOrder(c echo.Context) error {
	order := new(models.Order)

	if err := c.Bind(order); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Format JSON salah"})
	}

	if err := validate.Struct(order); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": err.Error()})
	}

	order.InvoiceNumber = "WPZ-" + time.Now().Format("20060102-150405")

	if order.PaymentMethod == "Cash" {
		order.Status = "Unpaid"
		if err := database.DB.Create(&order).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal simpan order"})
		}
		go sendEmailNotification(order, "")
		return c.JSON(http.StatusCreated, map[string]interface{}{
			"message": "✅ Pesanan dicatat. Silakan bayar di kasir.",
			"data": map[string]interface{}{
				"invoice_number": order.InvoiceNumber,
			},
		})
	}

	order.Status = "Pending Payment"
	if order.PaymentMethod == "" {
		order.PaymentMethod = "Digital"
	}

	if err := database.DB.Create(&order).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal simpan order"})
	}

	var s = snap.Client{}
	s.New(MidtransServerKey, midtrans.Sandbox)

	req := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  order.InvoiceNumber,
			GrossAmt: order.TotalPrice,
		},
		CustomerDetail: &midtrans.CustomerDetails{
			FName: order.CustomerName,
			Email: order.CustomerEmail,
		},
	}

	snapResp, err := s.CreateTransaction(req)
	if err != nil {
		fmt.Println("❌ ERROR MIDTRANS:", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Gagal menghubungi Payment Gateway",
			"error":   err.Error(),
		})
	}

	go sendEmailNotification(order, snapResp.RedirectURL)

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "✅ Order Dibuat! Silakan Bayar.",
		"data": map[string]interface{}{
			"invoice_number": order.InvoiceNumber,
			"payment_url":    snapResp.RedirectURL,
			"token":          snapResp.Token,
		},
	})
}

func GetAllOrders(c echo.Context) error {
	var orders []models.Order
	if err := database.DB.Order("created_at desc").Find(&orders).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Gagal mengambil data pesanan"})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": orders,
	})
}

// UpdateOrderStatus: DI SINI LOGIKA POTONG STOK & KIRIM PDF DITAMBAHKAN
func UpdateOrderStatus(c echo.Context) error {
	id := c.Param("id")
	var payload struct {
		Status string `json:"status"`
	}

	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Format JSON salah"})
	}

	var order models.Order
	if err := database.DB.Where("id = ?", id).First(&order).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "Pesanan tidak ditemukan"})
	}

	// Cek jika status berubah menjadi Paid
	if order.Status != "Paid" && payload.Status == "Paid" {
		go ReduceStock(order.ProductName) // Potong Stok
		go sendReceiptEmail(&order)       // Kirim PDF Lunas
	}

	order.Status = payload.Status
	database.DB.Save(&order)

	return c.JSON(http.StatusOK, map[string]string{"message": "Status pesanan berhasil diperbarui"})
}

// HandleMidtransNotification: DI SINI JUGA LOGIKA POTONG STOK & KIRIM PDF DITAMBAHKAN
func HandleMidtransNotification(c echo.Context) error {
	var notificationPayload map[string]interface{}

	if err := c.Bind(&notificationPayload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid JSON"})
	}

	orderID, exists := notificationPayload["order_id"].(string)
	transactionStatus, exists2 := notificationPayload["transaction_status"].(string)

	if !exists || !exists2 {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Data tidak lengkap"})
	}

	var newStatus string
	switch transactionStatus {
	case "capture", "settlement":
		newStatus = "Paid"
	case "deny", "cancel", "expire":
		newStatus = "Failed"
	case "pending":
		newStatus = "Pending"
	default:
		newStatus = "Unknown"
	}

	var order models.Order
	if err := database.DB.Where("invoice_number = ?", orderID).First(&order).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "Order tidak ditemukan"})
	}

	// Cek jika pembayaran berhasil (Paid)
	if order.Status != "Paid" && newStatus == "Paid" {
		go ReduceStock(order.ProductName) // Potong Stok
		go sendReceiptEmail(&order)       // Kirim PDF Lunas
	}

	order.Status = newStatus
	database.DB.Save(&order)

	fmt.Printf("✅ Update Order %s menjadi: %s\n", orderID, newStatus)

	return c.JSON(http.StatusOK, map[string]string{"message": "OK"})
}

// Fungsi ajaib untuk membaca string pesanan dan memotong stok di database
func deductProductStock(orderString string) {
	// Contoh orderString: "Indomie Goreng (2x) - Note: pedas | Es Teh (1x)"
	items := strings.Split(orderString, "|")
	
	// Regex untuk mengambil nama produk dan jumlah angka di dalam kurung (Nx)
	re := regexp.MustCompile(`^(.*?)\s\((\d+)x\)`)

	for _, item := range items {
		item = strings.TrimSpace(item)
		matches := re.FindStringSubmatch(item)
		
		if len(matches) >= 3 {
			productName := strings.TrimSpace(matches[1])
			qtyStr := matches[2]
			qty, _ := strconv.Atoi(qtyStr)

			// Perintah SQL menggunakan GORM untuk mengurangi stok berdasarkan nama menu
			err := database.DB.Exec("UPDATE products SET stock = stock - ? WHERE name = ?", qty, productName).Error
			if err != nil {
				log.Println("Gagal memotong stok untuk:", productName, err)
			} else {
				log.Printf("Stok %s berhasil dipotong sebanyak %d porsi\n", productName, qty)
			}
		}
	}
}