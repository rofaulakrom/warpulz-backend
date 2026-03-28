FROM golang:1.21-alpine
WORKDIR /app
COPY . .
RUN go build -o main .
EXPOSE 7860
ENV PORT=7860
CMD ["./main"]