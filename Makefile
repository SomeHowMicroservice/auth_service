ROOT := .
TMP_DIR := tmp
BIN := $(TMP_DIR)/main.exe
TESTDATA_DIR := testdata

.PHONY: build run clean

build:
	@echo "Building..."
	@if not exist $(TMP_DIR) mkdir $(TMP_DIR)
	go build -o $(BIN) .

run: build
	@echo "Running..."
	@$(BIN)

clean:
	@echo "Cleaning..."
	@rm -rf $(TMP_DIR)

proto-gen:
	@protoc --go_out=protobuf/auth --go-grpc_out=protobuf/auth \
	        proto/auth.proto
	@protoc --go_out=protobuf/user --go-grpc_out=protobuf/user \
	        proto/user.proto