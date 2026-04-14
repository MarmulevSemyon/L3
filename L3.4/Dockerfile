FROM golang:1.25.7 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN mkdir -p /app/bin && go build -o /app/bin/imageProcessor ./cmd/imageProcessor


FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/bin/imageProcessor /app/bin/imageProcessor
COPY --from=builder /app/configs /app/configs
COPY --from=builder /app/static /app/static
COPY --from=builder /app/uploads /app/uploads

EXPOSE 8080

CMD ["/app/bin/imageProcessor"]