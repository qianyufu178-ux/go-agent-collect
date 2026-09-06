.PHONY: all build clean test cpp cpp-test proto edge controller help

# Variables
GO := go
CGO_ENABLED := 1
GOOS := linux
GOARCH := amd64
BUILD_DIR := build
CPP_DIR := cpp
CPP_BUILD_DIR := $(BUILD_DIR)/cpp
EDGE_BIN := $(BUILD_DIR)/nexus-edge
CTRL_BIN := $(BUILD_DIR)/nexus-controller

# Default target
all: cpp proto edge controller

help:
	@echo "Usage:"
	@echo "  make all         - Build everything (cpp + proto + edge + controller)"
	@echo "  make cpp         - Build C++ static library"
	@echo "  make cpp-test    - Run C++ unit tests"
	@echo "  make proto       - Generate gRPC code from protobuf"
	@echo "  make edge        - Build Nexus-Edge binary"
	@echo "  make controller  - Build Nexus-Controller binary"
	@echo "  make build       - Build all binaries"
	@echo "  make test        - Run all tests"
	@echo "  make clean       - Clean build artifacts"

# C++ static library
cpp:
	@mkdir -p $(CPP_BUILD_DIR)
	cd $(CPP_BUILD_DIR) && cmake ../../$(CPP_DIR) -DCMAKE_BUILD_TYPE=Release && make -j$$(nproc)

cpp-test:
	@mkdir -p $(CPP_BUILD_DIR)
	cd $(CPP_BUILD_DIR) && cmake ../../$(CPP_DIR) -DCMAKE_BUILD_TYPE=Debug && make -j$$(nproc) && ctest --output-on-failure

# Protobuf
proto:
	@mkdir -p proto/gen
	protoc --go_out=proto/gen --go_opt=paths=source_relative \
		--go-grpc_out=proto/gen --go-grpc_opt=paths=source_relative \
		-I proto proto/nexus.proto

# Go binaries
edge: cpp proto
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) \
	$(GO) build -o $(EDGE_BIN) -ldflags="-s -w" ./edge/cmd

controller: cpp proto
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) \
	$(GO) build -o $(CTRL_BIN) -ldflags="-s -w" ./controller/cmd

build: edge controller

# Tests
test: cpp-test
	CGO_ENABLED=$(CGO_ENABLED) $(GO) test ./...

# Clean
clean:
	rm -rf $(BUILD_DIR)
	rm -rf proto/gen/*.go
	@echo "Clean completed"
