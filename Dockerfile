FROM golang:1.25.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /main ./cmd/api/main.go

FROM alpine:3.19

COPY --from=builder /main /main

COPY --from=builder /app/internal/db/migrations ./internal/db/migrations

EXPOSE ${PORT}

CMD ["./main"]