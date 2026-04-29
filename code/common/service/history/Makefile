.PHONY: start-api start-rpc test test-integration

API_DIR=./api
RPC_DIR=./rpc
SERVER=history
CONFIG=etc/$(SERVER).yaml

start-api:
	go run $(API_DIR)/*.go -f $(API_DIR)/$(CONFIG)

start-rpc:
	go run $(RPC_DIR)/*.go -f $(RPC_DIR)/$(CONFIG)

test:
	go test ./...

test-integration:
	GOCACHE=/tmp/go-build-history-integration GOTMPDIR=/tmp go test -tags=integration ./rpc/integration -run TestHistoryFlow_JSONTableDriven_Integration -v
