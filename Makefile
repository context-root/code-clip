.PHONY: all build run test install clean

# Default target
all: build

# Build the project
build:
	@mkdir -p bin
	go build -o bin/code-clip cmd/code-clip/main.go

# Run the project locally
run: build
	./bin/code-clip

# Run tests
test:
	go test ./... -v

# Install the binary locally into ~/go/bin, making it available in your PATH
install:
	go install ./cmd/code-clip

# Clean the target directory 
clean:
	rm -rf bin
	rm -f code-clip
