.PHONY: fmt vet test build clean

fmt:
	gofmt -w -s .

vet:
	go vet ./...

test:
	go test ./...

build:
	go build -o canarydrop ./cmd/canarydrop

clean:
	rm -f canarydrop


