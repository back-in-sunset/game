package cachex_test

import (
	"testing"
	"time"
	"user/utils/cachex"
)

func TestClockCache_Get(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		clocache *cachex.ClockCache[string]
		key      string
		want     any
		want2    bool
		setFunc  func(c *cachex.ClockCache[string])
	}{
		{
			name:     "get existing key",
			clocache: cachex.NewClockCache[string](10),
			key:      "key1",
			setFunc: func(c *cachex.ClockCache[string]) {
				c.Set("key1", "value1", time.Minute.Milliseconds())
			},
			want:  "value1",
			want2: true,
		},
		{
			name:  "get non-existing key",
			key:   "key2",
			want:  "",
			want2: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got2 := tt.clocache.Get(tt.key)
			if *got != tt.want {
				t.Errorf("Get() = %v, want %v", *got, tt.want)
			}
			if got2 != tt.want2 {
				t.Errorf("Get() = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func TestClockCache_GetInt(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		clocache *cachex.ClockCache[string]
		key      string
		want     any
		want2    bool
		setFunc  func(c *cachex.ClockCache[string])
	}{
		{
			name:     "get existing key",
			clocache: cachex.NewClockCache[string](10),
			key:      "key1",
			setFunc: func(c *cachex.ClockCache[string]) {
				c.Set("key1", "value1", time.Minute.Milliseconds())
			},
			want:  "value1",
			want2: true,
		},
		{
			name:  "get non-existing key",
			key:   "key2",
			want:  "",
			want2: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got2 := tt.clocache.Get(tt.key)
			if *got != tt.want {
				t.Errorf("Get() = %v, want %v", *got, tt.want)
			}
			if got2 != tt.want2 {
				t.Errorf("Get() = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func TestCacheMiss(t *testing.T) {
	// init cache
	cache := cachex.NewClockCache[string](3)

	cache.Set("k1", "v1", time.Minute.Milliseconds())
	cache.Set("k2", "v2", time.Minute.Milliseconds())
	cache.Set("k3", "v3", time.Minute.Milliseconds())
	// get k1
	cache.Get("k1")
	// get k2
	cache.Get("k2")
	// get k3
	cache.Get("k3")

	cache.Set("k4", "v4", time.Minute.Milliseconds())

	// get k1
	v1, ok1 := cache.Get("k1")
	if ok1 {
		t.Errorf("expected k1 to be present")
	}
	if *v1 != "" {
		t.Errorf("expected k1 to be v1, got %s", *v1)
	}

	// get k2
	cache.Get("k2")
	// get k2
	v2, ok2 := cache.Get("k2")
	if !ok2 {
		t.Errorf("expected k2 to be present")
	}
	if *v2 != "v2" {
		t.Errorf("expected k2 to be v2, got %s", *v2)
	}

	// get k3
	v3, ok3 := cache.Get("k3")
	if !ok3 {
		t.Errorf("expected k3 to be present")
	}
	if *v3 != "v3" {
		t.Errorf("expected k3 to be v3, got %s", *v3)
	}

	// get k4
	v4, ok4 := cache.Get("k4")
	if !ok4 {
		t.Errorf("expected k4 to be present")
	}
	if *v4 != "v4" {
		t.Errorf("expected k4 to be v4, got %s", *v4)
	}

}
