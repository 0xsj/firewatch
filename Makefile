.PHONY: fmt vet lint test coverage build docker-build clean ci

fmt:
	gofmt -w -s .

vet:
	go vet ./...

lint:
	golangci-lint run

test:
	go test -race ./...

coverage:
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out | tail -1

build:
	go build -trimpath -ldflags="-s -w" -o firewatch ./cmd/firewatch

docker-build:
	docker build -t firewatch:local .

clean:
	rm -f firewatch cmd/firewatch/firewatch coverage.out

ci: lint vet test build
