FROM golang:1.27.1-alpine3.24 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go install go.opentelemetry.io/otelc/tool/cmd/otelc@latest
RUN otelc go build -o myapi

FROM alpine:3.24
WORKDIR /app
COPY --from=builder /app/myapi .
RUN chmod +x ./myapi
EXPOSE 8888
CMD ["./myapi"]