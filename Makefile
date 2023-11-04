IMAGE_NAME := ghcr.io/lyqingye/autobangumi-go
TAG := latest
BINARY_NAME := autobangumi-go
OUTPUT_DIR := ./build/bin

.PHONY: build-docker build build-linux build-macos build-arm clean clean-local-deploy

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

.PHONY: local-clean
local-deploy: clean-local-deploy build-docker
	@mkdir ./deployment/db_data
	@mkdir -p aria2 qb downloads cache
	@docker compose -f ./deployment/docker-compose.yaml up -d

restart-local-deploy:
	@docker compose -f ./deployment/docker-compose.yaml restart

stop-local-deploy:
	@docker compose -f ./deployment/docker-compose.yaml down

clean-local-deploy:
	@docker compose -f ./deployment/docker-compose.yaml down
	@rm -rf ./deployment/aria2 ./deployment/qb ./deployment/db_data ./deployment/cache
