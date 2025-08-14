.PHONY: start rpc api gen model rpc-tag
 
API_DIR=./api
RPC_DIR=./rpc

SERVER=comment

CONFIG=etc/$(SERVER).yaml

start:

start-api:
	go run $(RPC_DIR)/*.go

start-rpc:
	go run $(RPC_DIR)/*.go -f $(RPC_DIR)/$(CONFIG)

gen: api rpc model
	echo "api rpc modol ok"

api:
	goctl api go -api $(API_DIR)/$(SERVER).api -dir $(API_DIR)

rpc:
	goctl rpc protoc $(RPC_DIR)/$(SERVER).proto --go_out=$(RPC_DIR) --go-grpc_out=$(RPC_DIR) --zrpc_out=$(RPC_DIR)
	
rpc-tag: rpc
	protoc-go-inject-tag -input="$(RPC_DIR)/$(SERVER)/*.pb.go" 

model:
	goctl model mysql ddl -s model/*.sql -d model -c 