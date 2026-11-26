package wal

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

// name=John exactly as it appears on disk:
// crc32 | keylen | vallen | deleted | key | val
var johnRecord = []byte{
	49, 52, 74, 199, // crc32 of the payload below
	4, 0, 0, 0, // key length
	4, 0, 0, 0, // value length
	0, // not deleted
	'n', 'a', 'm', 'e',
	'J', 'o', 'h', 'n',
}

// name tombstone: crc32 | keylen | vallen | deleted | key
var nameTombstone = []byte{
	190, 164, 202, 141, // crc32 of the payload below
	4, 0, 0, 0, // key length
	0, 0, 0, 0, // value length
	1, // deleted
	'n', 'a', 'm', 'e',
}

func TestEntryEncode(t *testing.T) {
	ent := Entry{
		Key: []byte("name"),
		Val: []byte("John"),
	}

	if got := ent.Encode(); !bytes.Equal(got, johnRecord) {
		t.Fatalf("Encode() = %v, want %v", got, johnRecord)
	}
}

func TestEntryDecode(t *testing.T) {
	var ent Entry

	err := ent.Decode(bytes.NewReader(johnRecord))
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
		t.Error("Deleted = true, want false")
	}
}

func TestEntryEncodeDeleted(t *testing.T) {
	ent := Entry{
		Key:     []byte("name"),
		Deleted: true,
	}

	if got := ent.Encode(); !bytes.Equal(got, nameTombstone) {
		t.Fatalf("Encode() = %v, want %v", got, nameTombstone)
	}
}

func TestEntryDecodeDeleted(t *testing.T) {
	var ent Entry

	err := ent.Decode(bytes.NewReader(nameTombstone))
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
		t.Error("Deleted = false, want true")
	}
}

func TestEntryRoundTrip(t *testing.T) {
	original := Entry{
		Key: []byte("username"),
		Val: []byte("abhinav"),
	}

	var decoded Entry

	if err := decoded.Decode(bytes.NewReader(original.Encode())); err != nil {
		t.Fatalf("Decode() returned error: %v", err)
	}

	if !bytes.Equal(original.Key, decoded.Key) {
		t.Errorf("key mismatch: got %s, want %s", decoded.Key, original.Key)
	}

	if !bytes.Equal(original.Val, decoded.Val) {
		t.Errorf("val mismatch: got %s, want %s", decoded.Val, original.Val)
	}

	if original.Deleted != decoded.Deleted {
		t.Errorf("Deleted mismatch: got %t, want %t", decoded.Deleted, original.Deleted)
	}
}

func TestEntryRoundTripDeleted(t *testing.T) {
	original := Entry{
		Key:     []byte("goner"),
		Deleted: true,
	}

	var decoded Entry

	if err := decoded.Decode(bytes.NewReader(original.Encode())); err != nil {
		t.Fatalf("Decode() returned error: %v", err)
	}

	if !bytes.Equal(original.Key, decoded.Key) {
		t.Errorf("key = %s, want %s", decoded.Key, original.Key)
	}

	if len(decoded.Val) != 0 {
		t.Errorf("val = %v, want empty", decoded.Val)
	}

	if !decoded.Deleted {
		t.Error("Deleted = false, want true")
	}
}

func TestEntryDecodeBadChecksum(t *testing.T) {
	encoded := bytes.Clone(johnRecord)
	encoded[0] ^= 0xff

	var ent Entry

	err := ent.Decode(bytes.NewReader(encoded))
	if !errors.Is(err, ErrBadSum) {
		t.Fatalf("Decode() error = %v, want %v", err, ErrBadSum)
	}
}

func TestEntryDecodeCorruptPayload(t *testing.T) {
	encoded := bytes.Clone(johnRecord)
	encoded[len(encoded)-1] ^= 0xff

	var ent Entry

	err := ent.Decode(bytes.NewReader(encoded))
	if !errors.Is(err, ErrBadSum) {
		t.Fatalf("Decode() error = %v, want %v", err, ErrBadSum)
	}
}

func TestEntryDecodeTruncated(t *testing.T) {
	truncated := johnRecord[:len(johnRecord)-3]

	var ent Entry

	err := ent.Decode(bytes.NewReader(truncated))
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("Decode() error = %v, want %v", err, io.ErrUnexpectedEOF)
	}
}
