FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/auth-service ./cmd/app

FROM alpine:3.18

WORKDIR /app

RUN apk add --no-cache curl && \
    curl -L https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /usr/local/bin/migrate && \
    chmod +x /usr/local/bin/migrate && \
    apk del curl

COPY --from=builder /app/auth-service .
COPY --from=builder /app/migration ./migration
COPY --from=builder /app/internal/app/config/local.env ./config/local.env

RUN adduser -D -g '' appuser
USER appuser

CMD ["./auth-service", "--env.path", "./config/local.env"]