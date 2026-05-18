FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o warehousecontrol ./cmd/warehousecontrol

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/warehousecontrol /app/warehousecontrol
COPY --from=builder /app/config /app/config
COPY --from=builder /app/web /app/web

EXPOSE 8080

CMD ["/app/warehousecontrol"]