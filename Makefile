IMAGE_NAME := lyqingye/autobangumi-go
TAG := latest
BINARY_NAME := autobangumi-go
OUTPUT_DIR := ./build/bin

.PHONY: build-docker build build-linux build-macos build-arm clean

build-docker:
	@docker build -t $(IMAGE_NAME):$(TAG) .

build:
	@go build -v -o $(OUTPUT_DIR)/$(BINARY_NAME) .

build-linux:
	@GOOS=linux GOARCH=amd64 go build -v -o $(OUTPUT_DIR)/$(BINARY_NAME)-linux .

build-macos:
	@GOOS=darwin GOARCH=amd64 go build -v -o $(OUTPUT_DIR)/$(BINARY_NAME)-macos .

build-arm:
	@GOOS=linux GOARCH=arm go build -v -o $(OUTPUT_DIR)/$(BINARY_NAME)-arm .

clean:
	@rm -rf $(OUTPUT_DIR)