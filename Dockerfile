FROM golang:1.24.0-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -trimpath -o /main ./cmd/api/main.go

# Local Development
FROM alpine:3.19 AS local
WORKDIR /app
RUN addgroup -g 1000 appuser && adduser -D -u 1000 -G appuser appuser
COPY --from=builder /main ./main
COPY --from=builder /app/internal/db/migrations ./internal/db/migrations
RUN chown -R appuser:appuser /app
USER appuser
CMD ["./main"]

# AWS Lambda Production
FROM alpine:3.19 AS lambda
WORKDIR /app
RUN addgroup -g 1000 appuser && adduser -D -u 1000 -G appuser appuser
# Install AWS Adapter
COPY --from=public.ecr.aws/awsguru/aws-lambda-adapter:0.8.4 /lambda-adapter /opt/extensions/lambda-adapter
COPY --from=builder /main /main
COPY --from=builder /app/internal/db/migrations ./internal/db/migrations
RUN chown -R appuser:appuser /app
USER appuser
CMD ["/main"]