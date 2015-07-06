package mitlru

import (
	"testing"
	"time"
)

func TestTTLRU(t *testing.T) {
	lru := NewTTLRUCache(5, 2*time.Second)

	lru.Add("key0", "val")
	lru.Add("key1", "val")
	lru.Add("key2", "val")
	lru.Add("key3", "val")
	lru.Add("key4", "val")

	if v, b := lru.Get("key0"); !b || v.(string) != "val" {
		t.Error("Should contain key0")
	}
	if v, b := lru.Get("key1"); !b || v.(string) != "val" {
		t.Error("Should contain key1")
	}
	if v, b := lru.Get("key2"); !b || v.(string) != "val" {
		t.Error("Should contain key2")
	}
	if v, b := lru.Get("key3"); !b || v.(string) != "val" {
		t.Error("Should contain key3")
	}
	if v, b := lru.Get("key4"); !b || v.(string) != "val" {
		t.Error("Should contain key4")
	}

	if lru.Len() != 5 {
		t.Error("Should contain 5, ", lru.Len())
	}

	lru.Add("key5", "val")

	if _, b := lru.Get("key0"); b {
		t.Error("Should not contain key0")
	}

	if v, b := lru.Get("key5"); !b || v.(string) != "val" {
		t.Error("Should contain key5")
	}

	lru.Get("key1")

	lru.Add("key6", "val")

	if v, b := lru.Get("key1"); !b || v.(string) != "val" {
		t.Error("Should contain key1")
	}

	if _, b := lru.Get("key2"); b {
		t.Error("Should not contain key2")
	}

	lru.Remove("key3")

	lru.Add("key7", "val")
	lru.Add("key6", "val")

	if _, b := lru.Get("key3"); b {
		t.Error("Should not contain key3")
	}

	if v, b := lru.Get("key4"); !b || v.(string) != "val" {
		t.Error("Should contain key4")
	}

	if lru.Capacity() != 5 {
		t.Error("Should hold 5")
	}

	lru.Purge()

	if lru.Len() != 0 {
		t.Error("Should contain 0")
	}

	if lru.Capacity() != 5 {
		t.Error("Should hold 5")
	}

	lru.Add("key7", "val")
	if lru.Len() != 1 {
		t.Error("Should contain 1")
	}

	time.Sleep(1 * time.Second)

	lru.Add("key8", "val")

	if lru.Len() != 2 {
		t.Error("Should contain 2")
	}

	time.Sleep(2 * time.Second)
	if lru.Len() >= 2 {
		t.Error("Should contain < 2")
	}

	time.Sleep(1 * time.Second)
	if lru.Len() != 0 {
		t.Error("Should contain 0")
	}
}
