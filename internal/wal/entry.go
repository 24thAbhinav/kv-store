package wal

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"
)

type Entry struct {
	Key     []byte
	Val     []byte
	Deleted bool
}

var ErrBadSum = errors.New("bad checksum")

// serialization
// checksum | keysize | valsize | deleted |  keydata | valdata
// 4 + 4 + 4 + 1  + len(key) + len(val)
func (ent *Entry) Encode() []byte {
	data := make([]byte, 4+4+4+1+len(ent.Key)+len(ent.Val))

	binary.LittleEndian.PutUint32(data[4:8], uint32(len(ent.Key)))
	binary.LittleEndian.PutUint32(data[8:12], uint32(len(ent.Val)))
	if ent.Deleted {
		data[12] = 1
	} else {
		data[12] = 0
	}

	copy(data[13:], ent.Key)
	copy(data[13+len(ent.Key):], ent.Val)
	checksum := crc32.ChecksumIEEE(data[4:])
	binary.LittleEndian.PutUint32(data[0:4], checksum)
	return data
}

func (ent *Entry) Decode(r io.Reader) error {

	checksumBuf := make([]byte, 4)

	if _, err := io.ReadFull(r, checksumBuf); err != nil {
		return err
	}

	keyLenBuf := make([]byte, 4)
	if _, err := io.ReadFull(r, keyLenBuf); err != nil {
		return err
	}

	valLenBuf := make([]byte, 4)
	if _, err := io.ReadFull(r, valLenBuf); err != nil {
		return err
	}

	deletedBuf := make([]byte, 1)
	if _, err := io.ReadFull(r, deletedBuf); err != nil {
		return err
	}
	ent.Deleted = deletedBuf[0] == 1

	keyLen := binary.LittleEndian.Uint32(keyLenBuf)
	valLen := binary.LittleEndian.Uint32(valLenBuf)

	ent.Key = make([]byte, keyLen)
	ent.Val = make([]byte, valLen)

	if _, err := io.ReadFull(r, ent.Key); err != nil {
		return err
	}

	if _, err := io.ReadFull(r, ent.Val); err != nil {
		return err
	}

	data := make([]byte, 9+len(ent.Key)+len(ent.Val))
	copy(data[0:4], keyLenBuf)
	copy(data[4:8], valLenBuf)
	copy(data[8:9], deletedBuf)
	copy(data[9:], ent.Key)
	copy(data[9+len(ent.Key):], ent.Val)
	expected := binary.LittleEndian.Uint32(checksumBuf)
	actual := crc32.ChecksumIEEE(data)

	if expected != actual {
		return ErrBadSum
	}

	return nil
}
