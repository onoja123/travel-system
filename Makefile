.PHONY: run build test clean docker-up docker-down

run:
	go run cmd/server/main.go

build:
	go build -o bin/server cmd/server/main.go

test:
	go test -v ./...

clean:
	rm -rf bin/

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

install:
	go mod download
	go mod tidy

migrate:
	go run cmd/migrate/main.go