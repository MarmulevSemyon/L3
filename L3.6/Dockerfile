FROM golang:1.25 AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /build/salestracker ./cmd/salestracker

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /build/salestracker /app/salestracker
COPY ./config /app/config
COPY ./web /app/web

EXPOSE 8080

CMD ["/app/salestracker"]