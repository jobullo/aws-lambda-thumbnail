# Variables
GOOS := linux
GOARCH := amd64
EXECUTABLE := bootstrap
ZIPFILE := function.zip

# Default target
all: clean build zip

# Build the Go executable
build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(EXECUTABLE) ./cmd/lambda

# Zip the executable
zip: build
	zip $(ZIPFILE) $(EXECUTABLE)

# Clean up generated files
clean:
	rm -f $(EXECUTABLE) $(ZIPFILE)

# Phony targets
.PHONY: all build zip clean
