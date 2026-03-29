FROM golang:1.25.6-alpine

# Install sertifikat keamanan & zona waktu (Penting untuk kirim Email Gmail & Midtrans!)
RUN apk update && apk add --no-cache git ca-certificates tzdata

# Matikan CGO agar proses kompilasi jauh lebih cepat dan ringan di server gratisan
ENV CGO_ENABLED=0 GOOS=linux

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build aplikasi
RUN go build -o main .

EXPOSE 7860
ENV PORT=7860

CMD ["./main"]