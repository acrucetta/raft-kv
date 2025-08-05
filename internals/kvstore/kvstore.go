package kvstore

import (
	"errors"
	"lsm-kv/internals/utils"
	"lsm-kv/internals/wal"
	"sync"

	"github.com/huandu/skiplist"
	"github.com/ian-kent/go-log/log"
)

const maxMemory = 10

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
	kv := &KVStore{
		list: skiplist.New(skiplist.String), // assumes string keys
		mu:   sync.RWMutex{},
		wal:  w,
	}

	// If the WAL contains information, add it to our log.
	entries, err := utils.ReadWAL(logPath)
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		switch entry.Command {
		case "SET":
			kv.list.Set(entry.Key, entry.Value)
		case "DELETE":
			kv.list.Set(entry.Key, Tombstone)
		}
	}
	return kv
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
	// TODO: Before writing to the memtable, check its size, if its
	// above a threshold, we want to flush it.
	if kv.list.Len() > maxMemory {
		// TODO: Flush the wal.
		log.Info("Flushing the memtable to an SST table.")
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
