# Use Debian-based image with CGO support
FROM golang:1.24

WORKDIR /app

# Enable CGO
ENV CGO_ENABLED=1
ENV GOOS=linux

# Install build dependencies for go-sqlite3
RUN apt-get update && apt-get install -y gcc libc6-dev

COPY . .

RUN go build -o api ./cmd/main.go

EXPOSE 8080
CMD ["./api"]
