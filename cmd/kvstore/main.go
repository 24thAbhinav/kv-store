package main

import (
	"fmt"
	"log"

	"kvstore/internal/kv"
)

func main() {
	store := &kv.KV{}
	if err := store.Open(); err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	const key = "name"

	store.Set([]byte(key), []byte("abhinav"))

	val, ok, err := store.Get([]byte(key))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s = %s (found: %t)\n", key, val, ok)

	deleted, err := store.Del([]byte(key))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("deleted %s: %t\n", key, deleted)
}
