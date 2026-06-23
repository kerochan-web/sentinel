# --- Build Stage ---
FROM golang:1.26-alpine AS builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o sentinel ./cmd/sentinel/main.go

# --- Final Run Stage ---
FROM alpine:3.21
RUN apk --no-cache add ca-certificates curl

# Install kubectl matching standard cluster API variants
RUN curl -LO "https://dl.k8s.io/release/v1.30.0/bin/linux/amd64/kubectl" \
    && chmod +x ./kubectl \
    && mv ./kubectl /usr/local/bin/kubectl

WORKDIR /app
COPY --from=builder /app/sentinel .

EXPOSE 2112
CMD ["./sentinel"]
