APP=shortener
CONTAINER=shortener-postgres
DB_USER=postgres
DB_PASSWORD=123
DB_NAME=shortener
DB_PORT=5433

.PHONY: fmt lint build test start db-up db-create migrate run stop clean

fmt:
	gofmt -w ./cmd ./internal ./pkg

lint:
	go vet ./...
	-golint ./...

build:
	mkdir -p bin
	go build -o bin/$(APP) ./cmd/shortener

test:
	go test ./...

db-up:
	@if docker ps -a --format '{{.Names}}' | grep -q '^$(CONTAINER)$$'; then \
		echo "Container $(CONTAINER) already exists"; \
		docker start $(CONTAINER) >/dev/null || true; \
	else \
		echo "Creating container $(CONTAINER)"; \
		docker run --name $(CONTAINER) \
			-e POSTGRES_USER=$(DB_USER) \
			-e POSTGRES_PASSWORD=$(DB_PASSWORD) \
			-e POSTGRES_DB=postgres \
			-p $(DB_PORT):5432 \
			-d postgres:16; \
	fi
	@echo "Waiting for PostgreSQL..."
	@until docker exec $(CONTAINER) pg_isready -U $(DB_USER) >/dev/null 2>&1; do \
		sleep 1; \
	done

db-create:
	@docker exec $(CONTAINER) psql -U $(DB_USER) -tc "SELECT 1 FROM pg_database WHERE datname = '$(DB_NAME)'" | grep -q 1 || \
	docker exec $(CONTAINER) psql -U $(DB_USER) -c "CREATE DATABASE $(DB_NAME)"

migrate:
	docker exec -i $(CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) < migrations/001_init.up.sql

start: fmt lint build test db-up db-create migrate
	./bin/$(APP)

run: start

stop:
	docker stop $(CONTAINER)

clean:
	rm -rf bin