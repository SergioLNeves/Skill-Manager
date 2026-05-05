.PHONY: dev build release-linux release-darwin test lint mocks clean

dev:
	wails dev

build:
	wails build

release-linux:
	wails build -clean -trimpath -ldflags "-s -w"
	tar -czf build/bin/skill-manager-linux-amd64.tar.gz -C build/bin skill-manager

release-darwin:
	wails build -clean -trimpath -platform darwin/universal -ldflags "-s -w"
	ditto -c -k --keepParent build/bin/skill-manager.app build/bin/skill-manager-darwin-universal.zip

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
