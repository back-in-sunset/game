.PHONY: build
build:
	cd rpc && go build -o bin/device_shadow rpc/*.go

.PHONY: run
run:
	cd rpc && go run rpc/*.go -f rpc/etc/deviceshadow.yaml

.PHONY: proto
proto:
	cd rpc && goctl rpc protoc device_shadow.proto --go_out=./device_shadow --go_opt=paths=source_relative --go-grpc_out=./device_shadow --go-grpc_opt=paths=source_relative --zrpc_out=.

.PHONY: tidy
tidy:
	cd rpc && go mod tidy

.PHONY: tidy-api
tidy-api:
	cd api && go mod tidy

.PHONY: test
test:
	cd rpc && go test -v ./...

.PHONY: clean
clean:
	rm -rf rpc/bin/
