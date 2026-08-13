# ---------- Build stage ----------
FROM golang:1.26.5 AS builder

WORKDIR /app

# Copy dependency files first
# This allows Docker to cache dependency downloads.
COPY go.mod go.sum ./

RUN go mod download

# Copy the application source
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o url-shortener ./cmd/url-shortener


# ---------- Runtime stage ----------
FROM alpine:3.22

WORKDIR /app

# CA certificates are useful for HTTPS requests.
RUN apk --no-cache add ca-certificates

# Copy only the compiled application
COPY --from=builder /app/url-shortener /app/url-shortener

EXPOSE 8082

CMD ["/app/url-shortener"]