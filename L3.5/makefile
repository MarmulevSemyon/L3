APP_NAME=eventBooker
BIN_DIR=bin
CMD_DIR=./cmd/eventBooker

.PHONY: build run clean vet lint compose-up compose-down

build: lint
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP_NAME) $(CMD_DIR)

run:
	go run $(CMD_DIR)
fmt:
	gofmt -w ./cmd ./internal

vet: fmt
	go vet ./...

lint: vet
	golint ./...


docker-up: lint
	docker compose up --build -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f

docker-restart:
	docker compose down -v
	docker compose up --build -d

clean:
	rm -rf $(BIN_DIR)
