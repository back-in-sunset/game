package cachex

// Cacher defines the interface for a cache with Get and Set methods.
type Cacher[T any] interface {
	Get(key string) (*T, bool)
	Set(key string, value T, ttlMs int64)
	Delete(key string)
}
