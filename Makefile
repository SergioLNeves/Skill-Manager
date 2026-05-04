.PHONY: dev build test lint mocks clean

dev:
	wails dev

build:
	wails build

test:
	go test ./...

test-verbose:
	go test -v -count=1 ./...

test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

lint:
	golangci-lint run ./...

mocks:
	mockery

generate:
	wails generate module

clean:
	rm -rf build/bin coverage.out
