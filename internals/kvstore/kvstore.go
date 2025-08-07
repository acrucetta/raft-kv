package kvstore

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"lsm-kv/internals/sstable"
	"lsm-kv/internals/utils"
	"lsm-kv/internals/wal"
	"os"
	"sort"
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

type EntryWithTimestamp struct {
	DirEntry  fs.DirEntry
	Timestamp int64
}

// Extracts the timestamp from the sst files. E.g., sst_2025_08_07_16:28:10.db
func parseTimestamp(name string) (int64, error) {
	// TODO: To implement.
	return 0, nil
}

func withTimestamps(entries []os.DirEntry) ([]EntryWithTimestamp, error) {
	out := make([]EntryWithTimestamp, 0)
	for _, entry := range entries {
		ts, err := parseTimestamp(entry.Name())
		if err != nil {
			log.Printf("Error parsing the timestamp: %v", err)
			return out, err
		}
		out = append(out, EntryWithTimestamp{entry, ts})
	}

	return out, nil
}

func lookupSSTables(key string) (string, error) {
	entries, err := os.ReadDir(sstable.SstDefaultPath)
	if err != nil {
		return "", fmt.Errorf("read sst dir: %w", err)
	}

	// Get all the available SST tables
	// Try to fetch the latest entry from newest to oldest.
	files, err := withTimestamps(entries)
	if err != nil {
		return "", err
	}

	// Sort the files
	sort.Slice(files, func(i, j int) bool {
		return files[i].Timestamp < files[j].Timestamp
	})

	// Find the key
	for _, f := range files {
		val, err := sstable.GetKey(&f.DirEntry)
		switch {
		case err == nil:
			return val, nil
		case errors.Is(err, sstable.ErrKeyNotFound):
			continue
		default:
			return "", fmt.Errorf("search %s: %w", f.DirEntry.Name(), err)
		}
	}
	return "", sstable.ErrKeyNotFound
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
	return lookupSSTables(key)
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
