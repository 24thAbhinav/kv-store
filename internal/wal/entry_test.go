package wal

import (
	"bytes"
	"testing"
)

func TestEntryEncode(t *testing.T) {
	ent := Entry{
		Key:     []byte("name"),
		Val:     []byte("John"),
		Deleted: false,
	}
	got := ent.Encode()

	expected := []byte{
		4, 0, 0, 0, // key length = 4 (little endian)
		4, 0, 0, 0, // value length = 4
		0,
		'n', 'a', 'm', 'e',
		'J', 'o', 'h', 'n',
	}

	if !bytes.Equal(got, expected) {
		t.Fatalf("Encode() = %v, want %v", got, expected)
	}
}

func TestEntryDecode(t *testing.T) {
	encoded := []byte{
		4, 0, 0, 0,
		4, 0, 0, 0,
		0,
		'n', 'a', 'm', 'e',
		'J', 'o', 'h', 'n',
	}

	reader := bytes.NewReader(encoded)

	var ent Entry

	err := ent.Decode(reader)
	if err != nil {
		t.Fatalf("Decode() returned error: %v", err)
	}

	if !bytes.Equal(ent.Key, []byte("name")) {
		t.Errorf("key = %s, want name", ent.Key)
	}

	if !bytes.Equal(ent.Val, []byte("John")) {
		t.Errorf("val = %s, want John", ent.Val)
	}

	if ent.Deleted {
		t.Error("deleted = true, want false")
	}
}

func TestEntryEncodeDeleted(t *testing.T) {
	ent := Entry{
		Key:     []byte("name"),
		Deleted: true,
	}
	got := ent.Encode()

	expected := []byte{
		4, 0, 0, 0, // key length = 4 (little endian)
		0, 0, 0, 0, // value length = 0
		1, // deleted
		'n', 'a', 'm', 'e',
	}

	if !bytes.Equal(got, expected) {
		t.Fatalf("Encode() = %v, want %v", got, expected)
	}
}

func TestEntryDecodeDeleted(t *testing.T) {
	encoded := []byte{
		4, 0, 0, 0,
		0, 0, 0, 0,
		1,
		'n', 'a', 'm', 'e',
	}

	reader := bytes.NewReader(encoded)

	var ent Entry

	err := ent.Decode(reader)
	if err != nil {
		t.Fatalf("Decode() returned error: %v", err)
	}

	if !bytes.Equal(ent.Key, []byte("name")) {
		t.Errorf("key = %s, want name", ent.Key)
	}

	if len(ent.Val) != 0 {
		t.Errorf("val = %v, want empty", ent.Val)
	}

	if !ent.Deleted {
		t.Error("deleted = false, want true")
	}
}

func TestEntryRoundTrip(t *testing.T) {
	original := Entry{
		Key:     []byte("username"),
		Val:     []byte("abhinav"),
		Deleted: false,
	}

	encoded := original.Encode()

	var decoded Entry

	reader := bytes.NewReader(encoded)

	err := decoded.Decode(reader)
	if err != nil {
		t.Fatalf("Decode() returned error: %v", err)
	}

	if !bytes.Equal(original.Key, decoded.Key) {
		t.Errorf("key mismatch: got %s, want %s",
			decoded.Key, original.Key)
	}

	if !bytes.Equal(original.Val, decoded.Val) {
		t.Errorf("value mismatch: got %s, want %s",
			decoded.Val, original.Val)
	}

	if original.Deleted != decoded.Deleted {
		t.Errorf("Deleted mismatch: got %t, want %t",
			decoded.Deleted, original.Deleted)
	}
}
