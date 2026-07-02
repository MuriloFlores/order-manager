FROM golang:1.22-alpine AS builder

WORKDIR /app

# Instala dependências do SO necessárias
RUN apk add --no-cache git

# Copia os arquivos de dependência
COPY go.mod go.sum ./
RUN go mod download

# Copia o código fonte
COPY . .

# Compila a aplicação
RUN CGO_ENABLED=0 GOOS=linux go build -o order-manager ./cmd/order-manager/main.go

# Imagem final mínima
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/order-manager .

CMD ["./order-manager"]
