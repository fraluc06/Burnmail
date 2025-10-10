.PHONY: build install clean run test

# Binary name
BINARY_NAME=burnmail

# Build the project
build:
	go build -o $(BINARY_NAME) -ldflags="-s -w" .

# Install to /usr/local/bin
install: build
	sudo mv $(BINARY_NAME) /usr/local/bin/

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-linux-amd64 -ldflags="-s -w" .
	GOOS=linux GOARCH=arm64 go build -o $(BINARY_NAME)-linux-arm64 -ldflags="-s -w" .
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY_NAME)-darwin-amd64 -ldflags="-s -w" .
	GOOS=darwin GOARCH=arm64 go build -o $(BINARY_NAME)-darwin-arm64 -ldflags="-s -w" .
	GOOS=windows GOARCH=amd64 go build -o $(BINARY_NAME)-windows-amd64.exe -ldflags="-s -w" .

# Clean build artifacts
clean:
	go clean
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-*

# Run the application
run:
	go run main.go

# Run tests
test:
	go test -v ./...

# Download dependencies
deps:
	go mod download
	go mod tidy
