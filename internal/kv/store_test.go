package kv

import (
	"bytes"
	"testing"
)

func newStore(t *testing.T) *KV {
	t.Helper()

	store := &KV{}
	if err := store.Open(); err != nil {
		t.Fatalf("Open() returned error: %v", err)
	}
	return store
}

func TestGetMissingKey(t *testing.T) {
	store := newStore(t)

	_, ok, err := store.Get([]byte("missing"))
	if err != ErrKeyNotFound {
		t.Fatalf("Get() error = %v, want %v", err, ErrKeyNotFound)
	}
	if ok {
		t.Fatal("Get() ok = true, want false")
	}
}

func TestSetAndGet(t *testing.T) {
	store := newStore(t)

	updated, err := store.Set([]byte("name"), []byte("abhinav"))
	if err != nil {
		t.Fatalf("Set() returned error: %v", err)
	}
	if updated {
		t.Error("Set() updated = true on new key, want false")
	}

	val, ok, err := store.Get([]byte("name"))
	if err != nil {
		t.Fatalf("Get() returned error: %v", err)
	}
	if !ok {
		t.Fatal("Get() ok = false, want true")
	}
	if !bytes.Equal(val, []byte("abhinav")) {
		t.Errorf("val = %s, want abhinav", val)
	}
}

func TestSetExistingKey(t *testing.T) {
	store := newStore(t)

	store.Set([]byte("name"), []byte("first"))

	updated, err := store.Set([]byte("name"), []byte("second"))
	if err != nil {
		t.Fatalf("Set() returned error: %v", err)
	}
	if !updated {
		t.Error("Set() updated = false on existing key, want true")
	}

	val, _, _ := store.Get([]byte("name"))
	if !bytes.Equal(val, []byte("second")) {
		t.Errorf("val = %s, want second", val)
	}
}

func TestDel(t *testing.T) {
	store := newStore(t)

	store.Set([]byte("name"), []byte("abhinav"))

	deleted, err := store.Del([]byte("name"))
	if err != nil {
		t.Fatalf("Del() returned error: %v", err)
	}
	if !deleted {
		t.Error("Del() = false on existing key, want true")
	}

	if _, _, err := store.Get([]byte("name")); err != ErrKeyNotFound {
		t.Errorf("Get() after Del() error = %v, want %v", err, ErrKeyNotFound)
	}

	deleted, err = store.Del([]byte("name"))
	if err != nil {
		t.Fatalf("Del() returned error: %v", err)
	}
	if deleted {
		t.Error("Del() = true on missing key, want false")
	}
}
