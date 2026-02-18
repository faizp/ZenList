FROM golang:1.24.0-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/api ./cmd/api && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/migrate ./cmd/migrate

FROM alpine:3.20
WORKDIR /app

RUN adduser -D -H appuser

COPY --from=builder /out/api /usr/local/bin/api
COPY --from=builder /out/migrate /usr/local/bin/migrate
COPY --from=builder /app/migrations /app/migrations

USER appuser
EXPOSE 8080

CMD ["sh", "-c", "migrate -direction up && api"]
