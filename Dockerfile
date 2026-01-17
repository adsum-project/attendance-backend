# Stage 1: Build
FROM golang:1.23-alpine AS builder
WORKDIR /app

# Download modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build specific file, but output as "main"
RUN go build -o main ./cmd/attendanceapi/main.go

# Stage 2: Run
FROM alpine:latest
WORKDIR /root/

# Copy the binary named "main"
COPY --from=builder /app/main .

# Expose port
EXPOSE 8080

# Run it
CMD ["./main"]