# UI Dockerfile
FROM golang:1.24-alpine

WORKDIR /app
COPY . .

RUN go build -o ui ./cmd/main.go

EXPOSE 8081
CMD ["./ui"]
