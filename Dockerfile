# Step 1: Build the binary in a Go environment
FROM golang:1.23-alpine AS builder
WORKDIR /app
# Copy dependencies first (optimizes Docker caching)
COPY go.mod go.sum ./
RUN go mod download
# Copy source and build
COPY . .
RUN go build -o storagevault .

# Step 2: Use a tiny image for production (Security/Speed)
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/storagevault .
# Create the directory your store.go expects
RUN mkdir -p /root/network_storage 
EXPOSE 3000
CMD ["./storagevault"]