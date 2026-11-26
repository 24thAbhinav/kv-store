package wal

import (
	"bytes"
	"path/filepath"
	"testing"
)

func walPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "test.wal")
}

func openLog(t *testing.T, path string) *Log {
	t.Helper()

	log := &Log{FileName: path}
	if err := log.Open(); err != nil {
		t.Fatalf("Open() returned error: %v", err)
	}
	return log
}

func TestLogReadEmpty(t *testing.T) {
	log := openLog(t, walPath(t))
	defer log.Close()

	var ent Entry
	eof, err := log.Read(&ent)
	if err != nil {
		t.Fatalf("Read() returned error: %v", err)
	}
	if !eof {
		t.Error("Read() eof = false on empty log, want true")
	}
}

func TestLogWriteThenRead(t *testing.T) {
	path := walPath(t)

	log := openLog(t, path)
	want := []Entry{
		{Key: []byte("a"), Val: []byte("1")},
		{Key: []byte("b"), Val: []byte("2")},
	}
	for i := range want {
		if err := log.Write(&want[i]); err != nil {
			t.Fatalf("Write() returned error: %v", err)
		}
	}
	if err := log.Close(); err != nil {
		t.Fatalf("Close() returned error: %v", err)
	}

	read := openLog(t, path)
	defer read.Close()

	for i := range want {
		var got Entry
		eof, err := read.Read(&got)
		if err != nil {
			t.Fatalf("Read() returned error: %v", err)
		}
		if eof {
			t.Fatalf("Read() eof = true on entry %d, want false", i)
		}
		if !bytes.Equal(got.Key, want[i].Key) || !bytes.Equal(got.Val, want[i].Val) {
			t.Errorf("entry %d = %s=%s, want %s=%s",
				i, got.Key, got.Val, want[i].Key, want[i].Val)
		}
	}

	var ent Entry
	eof, err := read.Read(&ent)
	if err != nil {
		t.Fatalf("Read() at end returned error: %v", err)
	}
	if !eof {
		t.Error("Read() eof = false at end of log, want true")
	}
}

func TestLogAppendAfterReopen(t *testing.T) {
	path := walPath(t)

	first := openLog(t, path)
	if err := first.Write(&Entry{Key: []byte("a"), Val: []byte("1")}); err != nil {
		t.Fatalf("Write() returned error: %v", err)
	}
	if err := first.Close(); err != nil {
		t.Fatalf("Close() returned error: %v", err)
	}

	second := openLog(t, path)
	if err := second.Write(&Entry{Key: []byte("b"), Val: []byte("2")}); err != nil {
		t.Fatalf("Write() after reopen returned error: %v", err)
	}
	if err := second.Close(); err != nil {
		t.Fatalf("Close() returned error: %v", err)
	}

	read := openLog(t, path)
	defer read.Close()

	for _, want := range []string{"a", "b"} {
		var got Entry
		eof, err := read.Read(&got)
		if err != nil {
			t.Fatalf("Read() returned error: %v", err)
		}
		if eof {
			t.Fatalf("Read() eof = true, want entry with key %s", want)
		}
		if string(got.Key) != want {
			t.Errorf("key = %s, want %s", got.Key, want)
		}
	}
}
