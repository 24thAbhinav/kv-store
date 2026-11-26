package kv

import (
	"bytes"
	"io"
	"os"
	"testing"

	"kvstore/internal/wal"
)

func newStore(t *testing.T) *KV {
	t.Helper()

	t.Chdir(t.TempDir())

	store := &KV{}
	if err := store.Open(); err != nil {
		t.Fatalf("Open() returned error: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	return store
}

func readLog(t *testing.T, path string) []wal.Entry {
	t.Helper()

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("Open(%q) returned error: %v", path, err)
	}
	defer f.Close()

	var entries []wal.Entry
	for {
		var ent wal.Entry
		err := ent.Decode(f)
		if err == io.EOF {
			return entries
		}
		if err != nil {
			t.Fatalf("Decode() returned error: %v", err)
		}
		entries = append(entries, ent)
	}
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

func TestSetWritesLogEntry(t *testing.T) {
	store := newStore(t)

	if _, err := store.Set([]byte("name"), []byte("abhinav")); err != nil {
		t.Fatalf("Set() returned error: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("Close() returned error: %v", err)
	}

	entries := readLog(t, "store.wal")
	if len(entries) != 1 {
		t.Fatalf("log has %d entries, want 1", len(entries))
	}

	ent := entries[0]
	if string(ent.Key) != "name" || string(ent.Val) != "abhinav" {
		t.Errorf("entry = %s=%s, want name=abhinav", ent.Key, ent.Val)
	}
	if ent.Deleted {
		t.Error("entry Deleted = true, want false")
	}
}

func TestDelWritesTombstone(t *testing.T) {
	store := newStore(t)

	store.Set([]byte("name"), []byte("abhinav"))

	if _, err := store.Del([]byte("name")); err != nil {
		t.Fatalf("Del() returned error: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("Close() returned error: %v", err)
	}

	entries := readLog(t, "store.wal")
	if len(entries) != 2 {
		t.Fatalf("log has %d entries, want 2", len(entries))
	}

	tombstone := entries[1]
	if !tombstone.Deleted {
		t.Error("tombstone Deleted = false, want true")
	}
	if string(tombstone.Key) != "name" {
		t.Errorf("tombstone key = %s, want name", tombstone.Key)
	}
	if len(tombstone.Val) != 0 {
		t.Errorf("tombstone val = %v, want empty", tombstone.Val)
	}
}

func TestDelMissingWritesNothing(t *testing.T) {
	store := newStore(t)

	deleted, err := store.Del([]byte("missing"))
	if err != nil {
		t.Fatalf("Del() returned error: %v", err)
	}
	if deleted {
		t.Error("Del() = true on missing key, want false")
	}
	if err := store.Close(); err != nil {
		t.Fatalf("Close() returned error: %v", err)
	}

	entries := readLog(t, "store.wal")
	if len(entries) != 0 {
		t.Errorf("log has %d entries, want 0", len(entries))
	}
}
