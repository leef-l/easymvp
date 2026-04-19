ROOT_DIR := $(shell pwd)
APP_NAME := easymvp-core
BIN_DIR := $(ROOT_DIR)/bin
BIN_PATH := $(BIN_DIR)/$(APP_NAME)

.PHONY: test
test:
	go test ./...

.PHONY: build
build:
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 go build -o $(BIN_PATH) ./main.go

.PHONY: build-linux
build-linux:
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(BIN_DIR)/$(APP_NAME)-linux-amd64 ./main.go

.PHONY: build-windows
build-windows:
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o $(BIN_DIR)/$(APP_NAME)-windows-amd64.exe ./main.go

.PHONY: docker-build
docker-build:
	docker build -f manifest/docker/Dockerfile -t $(APP_NAME):local .

.PHONY: clean
clean:
	rm -rf $(BIN_DIR)
