package kvstore

import (
	"log"
	"lsm-kv/internals/sstable"
	"lsm-kv/internals/utils"
	"lsm-kv/internals/wal"
	"sync"

	"github.com/huandu/skiplist"
)

const maxMemory = 10

var Tombstone = "__<deleted>__"

type KVStore struct {
	list *skiplist.SkipList
	mu   sync.RWMutex
	wal  *wal.Wal
}

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

	// Before writing to the memtable, check its size, if its
	// above a threshold, we want to flush it.
	if kv.list.Len() > maxMemory {
		// Flush current memtable to the database
		filename, err := sstable.FlushToSSTable(kv.list)
		if err != nil {
			log.Printf("Error flushing to SST: %v", err)
			return err
		}
		log.Printf("Flushing the memtable to an SST table: %v", filename)
		// Now we delete the existing WAL (or truncate it)
		err = kv.wal.File.Truncate(0)
		if err != nil {
			log.Printf("Error truncating wal: %v", err)
			return err
		}
		_, err = kv.wal.File.Seek(0, 0)
		if err != nil {
			log.Printf("Error moving the seek to 0 in the wal: %v", err)
			return err
		}
	}
	kv.list.Set(key, value)
	return err
}

func (kv *KVStore) Get(key string) (string, error) {
	kv.mu.RLock()
	defer kv.mu.RUnlock()
	v, ok := kv.list.GetValue(key)

	// First check the Memtable.
	if ok && v != Tombstone {
		return v.(string), nil
	}
	// Second, search SST tables (from newest to oldest)
	return sstable.GetKeyFromSSTables(key)
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
	sstable.WriteSSTableEntry(key, Tombstone)
	if err != nil {
		return err
	}
	return nil
}
