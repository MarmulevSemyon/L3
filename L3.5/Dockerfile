
FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app/bin/eventBooker ./cmd/eventBooker


FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/bin/eventBooker /app/eventBooker
COPY ./web /app/web
COPY ./migrations /app/migrations
COPY ./config /app/config

EXPOSE 8080

CMD ["/app/eventBooker"]