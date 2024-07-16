BINARY_NAME=mycli-extension
VERSION=$(shell git describe --tags --always --long --dirty)

.PHONY: build
build:
    go build -o $(BINARY_NAME) -ldflags "-X main.Version=$(VERSION)"

.PHONY: test
test:
    go test ./...

.PHONY: clean
clean:
    go clean
    rm -f $(BINARY_NAME)

.PHONY: release
release:
    goreleaser release
