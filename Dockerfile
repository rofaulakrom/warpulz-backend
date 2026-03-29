# Menggunakan versi Go yang lebih baru (1.23) agar kompatibel
FROM golang:1.23-alpine

WORKDIR /app

# Install git jika diperlukan untuk mengambil dependency
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build aplikasi
RUN go build -o main .

# Hugging Face menggunakan port 7860 secara default
EXPOSE 7860
ENV PORT=7860

CMD ["./main"]