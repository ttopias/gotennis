BINARY_NAME = goTennis

all: clean deps lint test build

clean:
	@echo "Cleaning up..."
	@rm -f $(BINARY_NAME)

deps:
	@echo "Installing dependencies..."
	go mod download

lint:
	@echo "Linting..."
	go fmt ./...
	golangci-lint run -v ./...

test:
	@echo "Running tests..."
	go test -v -race ./...

build:
	@echo "Building..."
	go build -o $(BINARY_NAME) -v
	./$(BINARY_NAME)

install:
	@echo "Installing..."
	go install

run:
	@echo "Running..."
	go build -o $(BINARY_NAME) -v ./...
	./$(BINARY_NAME)

.PHONY: all clean deps lint test build install run