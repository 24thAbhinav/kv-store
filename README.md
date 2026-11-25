# KV Store

A small key-value store I'm building to learn how storage engines work, mostly
following the Bitcask paper. So far it's an in-memory map plus the entry
encoding. The append-only log comes next.

## Layout

- `cmd/kvstore` — entry point; runs a quick set/get/del so there's something to run
- `internal/kv` — the store itself: `Open`, `Get`, `Set`, `Del` over a map
- `internal/wal` — `Entry` and its on-disk encoding

## Entry format

Every entry is written as:

```
keysize (4 bytes) | valsize (4 bytes) | key | val
```

Sizes are little-endian `uint32`. `Decode` reads from any `io.Reader`, so the
same code will work once entries are read back from a file.

## Run

```
go run ./cmd/kvstore
```

## Test

```
go test ./...
```

## Roadmap

- [x] in-memory store
- [x] entry encoding and decoding
- [ ] append-only log: append entries to a file, replay them on open
- [ ] key -> offset index so reads don't scan the whole log
- [ ] compaction so the log doesn't grow forever
- [ ] hint files
- [ ] checksums and torn-write handling
