.PHONY: start rpc gen model test test-logic

RPC_DIR=./rpc
MODEL_DIR=./model

SERVER=friend

CONFIG=etc/$(SERVER).yaml

start:
	go run $(RPC_DIR)/*.go -f $(RPC_DIR)/$(CONFIG)

gen: rpc model
	echo "rpc model ok"

rpc:
	goctl rpc protoc $(RPC_DIR)/$(SERVER).proto --go_out=$(RPC_DIR) --go-grpc_out=$(RPC_DIR) --zrpc_out=$(RPC_DIR)

model:
	goctl model mysql ddl -s $(MODEL_DIR)/*.sql -d model -c

test:
	go test ./...

test-logic:
	go test ./rpc/internal/logic
