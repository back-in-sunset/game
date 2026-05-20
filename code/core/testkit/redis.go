package testkit

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
)

func StartMiniRedis(t *testing.T) string {
	t.Helper()

	srv, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run() error = %v", err)
	}
	t.Cleanup(srv.Close)
	return srv.Addr()
}
