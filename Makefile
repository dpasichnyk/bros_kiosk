.PHONY: up test lint fmt coverage build clean

up:
	go run cmd/server/main.go

build:
	go build -o bin/server cmd/server/main.go

build-pi:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=6 go build -ldflags="-s -w" -o bin/server-pi cmd/server/main.go

test:
	go test -v ./...

lint:
	golangci-lint run ./...

fmt:
	go fmt ./...

coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

clean:
	rm -rf bin/
	rm -f coverage.out
