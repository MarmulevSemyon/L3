APP_NAME=commentTree
BIN_DIR=bin
CMD_DIR=./cmd/commentTree

.PHONY: build run clean vet lint compose-up compose-down

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP_NAME) $(CMD_DIR)

run:
	go run $(CMD_DIR)
fmt:
	gofmt -w ./cmd ./internal

vet: fmt
	go vet ./...

lint: fmt vet
	golint run ./...


docker-up: vet build
	docker compose up --build -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f

docker-restart:
	docker compose down
	docker compose up --build -d

clean:
	rm -rf $(BIN_DIR)
