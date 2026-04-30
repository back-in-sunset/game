.PHONY: test test-integration

test:
	go test ./...

test-integration:
	GOCACHE=/tmp/go-build-im-integration GOTMPDIR=/tmp go test -tags=integration ./integration -run TestMySQLRedisStore_Integration -v
