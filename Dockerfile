FROM golang:1.24.0-alpine AS base
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
FROM base AS builder
# Build the binary
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /main ./cmd/api/main.go

# Local Development
FROM alpine:3.19 AS local
WORKDIR /app
COPY --from=builder /main /main
COPY --from=builder /app/internal/db/migrations ./internal/db/migrations
CMD ["./main"]

# AWS Lambda Production
FROM alpine:3.19 AS lambda
WORKDIR /app
# Install AWS Adapter
COPY --from=public.ecr.aws/awsguru/aws-lambda-adapter:0.8.4 /lambda-adapter /opt/extensions/lambda-adapter
COPY --from=builder /main /main
COPY --from=builder /app/internal/db/migrations ./internal/db/migrations
CMD ["/main"]