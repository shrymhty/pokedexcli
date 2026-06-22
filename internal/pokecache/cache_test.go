package pokecache

import (
	"testing"
	"time"
)

func TestAddGet(t *testing.T) {
	cache := NewCache(5 * time.Second)

	key := "https://example.com"
	val := []byte("data")

	cache.Add(key, val)

	got, ok := cache.Get(key)

	if !ok {
		t.Errorf("Key not found. Expected to find key")
	}

	if string(got) != string(val) {
        t.Errorf("expected %s, got %s", val, got)
    }
}

func TestReapLoop(t *testing.T) {
	interval := time.Millisecond * 10
	cache := NewCache(interval)

	key := "https://example.com"
	val := []byte("data")

	cache.Add(key, val)

	time.Sleep(interval + time.Millisecond*20)

	_, ok := cache.Get(key)

	if ok {
		t.Errorf("expected key to be reaped, but it was found")
	}
}