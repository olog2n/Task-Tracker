# Build stage
FROM golang:1.26-alpine AS build
WORKDIR /app

ENV GOPROXY=https://proxy.golang.org,direct
RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o /app/server ./cmd/server/main.go

# Production stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app

# 👇 Создай директорию configs
RUN mkdir -p /app/configs

COPY --from=build /app/server .

# 👇 Копируй конфиг ПРЯМО В ОБРАЗ (не через volume!)
# COPY ./config.docker.yaml ./app/configs/config.yaml

EXPOSE 8080
CMD ["./server"]