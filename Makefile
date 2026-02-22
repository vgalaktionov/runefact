BINARY := runefact
BUILD_DIR := bin

.PHONY: build run test lint vet clean

build:
	go build -o $(BUILD_DIR)/$(BINARY) ./cmd/runefact

run:
	go run ./cmd/runefact $(ARGS)

test:
	go test ./...

lint: vet

vet:
	go vet ./...

clean:
	rm -rf $(BUILD_DIR)
