.PHONY: all server client clean up

# Go build settings
GO ?= CGO_ENABLED=0 go
SERVER_SRC := ./cmd/server
CLIENT_SRC := ./cmd/client
BIN_DIR := bin

build-all: build-server build-client

build-server:
	@mkdir -p $(BIN_DIR)
	$(GO) build -v -o $(BIN_DIR)/server $(SERVER_SRC)

build-client:
	@mkdir -p $(BIN_DIR)
	$(GO) build -v -o $(BIN_DIR)/client $(CLIENT_SRC)

clean:
	rm -rf $(BIN_DIR)

up:
	docker compose up -d --build

down:
	docker compose down -v

logs:	
	docker compose logs -f