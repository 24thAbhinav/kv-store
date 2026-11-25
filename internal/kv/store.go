package kv

import (
	"errors"
	"kvstore/internal/wal"
)

var ErrKeyNotFound = errors.New("key not found")

type KV struct {
	log wal.Log
	mem map[string][]byte
}

func (kv *KV) Open() error {
	kv.mem = map[string][]byte{}

	kv.log = wal.Log{FileName: "store.wal"}

	if err := kv.log.Open(); err != nil {
		return err
	}

	return nil
}

func (kv *KV) Close() error {
	if err := kv.log.Close(); err != nil {
		return err
	}
	return nil
}

func (kv *KV) Get(key []byte) (val []byte, ok bool, err error) {
	value, success := kv.mem[string(key)]

	if !success {
		return nil, false, ErrKeyNotFound
	}

	return value, true, nil
}

func (kv *KV) Set(key []byte, val []byte) (updated bool, err error) {
	_, updated = kv.mem[string(key)]
	ent := &wal.Entry{
		Key: key,
		Val: val,
	}

	if err := kv.log.Write(ent); err != nil {
		return false, err
	}

	kv.mem[string(key)] = val

	return updated, nil
}

func (kv *KV) Del(key []byte) (deleted bool, err error) {
	_, deleted = kv.mem[string(key)]
	if !deleted {
		return false, nil
	}

	entry := wal.Entry{
		Key:     key,
		Deleted: true,
	}

	if err := kv.log.Write(&entry); err != nil {
		return false, err
	}
	delete(kv.mem, string(key))

	return deleted, nil
}
