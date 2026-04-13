FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o commentTree ./cmd/commentTree

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/commentTree /app/commentTree
COPY --from=builder /app/configs /app/configs
COPY --from=builder /app/static /app/static
COPY --from=builder /app/migrations /app/migrations

EXPOSE 8080

CMD ["/app/commentTree"]