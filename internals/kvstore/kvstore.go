package kvstore

import (
	"errors"
	"lsm-kv/internals/wal"
	"sync"

	"github.com/huandu/skiplist"
)

type KVStore struct {
	list *skiplist.SkipList
	mu   sync.RWMutex
	wal  *wal.Wal
}

var Tombstone = "__<deleted>__"

func NewKVStore(logPath string) *KVStore {
	w, err := wal.NewLogFile(logPath)
	if err != nil {
		panic(err) // or handle the error as appropriate
	}
	return &KVStore{
		list: skiplist.New(skiplist.String), // assumes string keys
		mu:   sync.RWMutex{},
		wal:  w,
	}
}

func (kv *KVStore) Set(key string, value string) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	entry := wal.LogEntry{
		Command: "SET",
		Key:     key,
		Value:   value,
	}
	err := kv.wal.AppendEntry(entry)
	if err != nil {
		return err
	}
	kv.list.Set(key, value)
	return err
}

func (kv *KVStore) Get(key string) (string, error) {
	kv.mu.RLock()
	defer kv.mu.RUnlock()
	v, ok := kv.list.GetValue(key)
	if !ok || v == Tombstone {
		return "", errors.New("key not found")
	}
	return v.(string), nil
}

func (kv *KVStore) Delete(key string) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	entry := wal.LogEntry{
		Command: "DELETE",
		Key:     key,
		Value:   Tombstone,
	}
	err := kv.wal.AppendEntry(entry)
	if err != nil {
		return err
	}
	kv.list.Set(key, Tombstone)
	return nil
}
