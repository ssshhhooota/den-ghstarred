BINARY := den-ghstarred

.PHONY: build test lint clean install run fmt

build:
	go build -o $(BINARY) .

test:
	go test ./...

lint:
	golangci-lint run ./...

clean:
	rm -f $(BINARY)

install:
	go install .

run:
	go run .

fmt:
	goimports -w .
	gofmt -w .
