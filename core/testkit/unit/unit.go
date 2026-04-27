package unit

import (
	"context"
	"testing"
)

func ContextWithValue(key, value any) context.Context {
	return context.WithValue(context.Background(), key, value)
}

func RequireNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func RequireEqual[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got != want {
		t.Fatalf("got = %v, want = %v", got, want)
	}
}
