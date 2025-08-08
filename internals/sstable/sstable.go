package sstable

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"lsm-kv/internals/utils"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/huandu/skiplist"
)

var ErrKeyNotFound = errors.New("kv: key not found")

type SSTEntry struct {
	KeyLength   int
	Key         string
	ValueLength int
	Value       string
}

type EntryWithTimestamp struct {
	DirEntry  fs.DirEntry
	Timestamp int64
}

const SstDefaultPath = "../../sst"

func FlushToSSTable(list *skiplist.SkipList) (string, error) {
	ts := time.Now().UTC().Format("2006_01_02_15:04:05")
	filename := fmt.Sprintf("sst_%v.db", ts)
	fullPath := filepath.Join(SstDefaultPath, filename)
	file, err := os.OpenFile(fullPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	for elem := list.Front(); elem != nil; elem = elem.Next() {
		key := []byte(elem.Key().(string))
		val := []byte(elem.Value.(string))

		// Format: [key len][key][val len][val]
		binary.Write(file, binary.LittleEndian, int32(len(key)))
		file.Write(key)
		binary.Write(file, binary.LittleEndian, int32(len(val)))
		file.Write(val)
	}
	return filename, nil
}

// Finds the latest SSTable and writes an entry to it
// at the end of the file.
func WriteSSTableEntry(key string, value string) error {
	files, err := getSortedSSTables()
	if err != nil {
		return err
	}
	f := files[0]
	path := filepath.Join(SstDefaultPath, f.DirEntry.Name())
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	keyBytes := []byte(key)
	valueBytes := []byte(value)

	// Format: [key len][key][val len][val]
	binary.Write(file, binary.LittleEndian, int32(len(key)))
	file.Write(keyBytes)
	binary.Write(file, binary.LittleEndian, int32(len(value)))
	file.Write(valueBytes)
	return nil
}

func ReadSSTTableEntry(file *os.File) (key string, value string, err error) {
	// Read Key Len
	var keyLen int32
	err = binary.Read(file, binary.LittleEndian, &keyLen)
	if err != nil {
		return "", "", err
	}
	// Read Key
	keyBytes := make([]byte, keyLen)
	err = binary.Read(file, binary.LittleEndian, &keyBytes)
	if err != nil {
		return "", "", err
	}
	// Read Value Len
	var valueLen int32
	err = binary.Read(file, binary.LittleEndian, &valueLen)
	if err != nil {
		return "", "", err
	}
	// Read Value
	valueBytes := make([]byte, valueLen)
	err = binary.Read(file, binary.LittleEndian, &valueBytes)
	if err != nil {
		return "", "", err
	}
	return string(keyBytes), string(valueBytes), nil
}

func FetchKeyValueInSSTable(fileName string, searchKey string) (string, bool, error) {
	filePath := filepath.Join(SstDefaultPath, fileName)
	file, err := os.Open(filePath)
	if err != nil {
		return "", false, fmt.Errorf("could not open SST file: %w", err)
	}
	defer file.Close()
	for {
		key, value, err := ReadSSTTableEntry(file)
		if err == io.EOF {
			return "", false, ErrKeyNotFound
		}
		if err != nil {
			return "", false, err
		}
		if key == searchKey {
			return value, true, nil
		}
	}
}

func withTimestamps(entries []os.DirEntry) ([]EntryWithTimestamp, error) {
	out := make([]EntryWithTimestamp, 0)
	for _, entry := range entries {
		ts, err := utils.ParseTimestamp(entry.Name())
		if err != nil {
			log.Printf("Error parsing the timestamp: %v", err)
			return out, err
		}
		out = append(out, EntryWithTimestamp{entry, ts})
	}
	return out, nil
}

func getSortedSSTables() ([]EntryWithTimestamp, error) {
	dirEntries, err := os.ReadDir(SstDefaultPath)
	if err != nil {
		return nil, fmt.Errorf("read sst dir: %w", err)
	}

	// Get all the available SST tables
	// Try to fetch the latest entry from newest to oldest.
	files, err := withTimestamps(dirEntries)
	if err != nil {
		return nil, err
	}

	// Sort the files
	sort.Slice(files, func(i, j int) bool {
		return files[i].Timestamp < files[j].Timestamp
	})
	return files, nil
}

func GetKeyFromSSTables(key string) (string, error) {
	// Find the key
	files, err := getSortedSSTables()
	if err != nil {
		return "", err
	}
	for _, f := range files {
		val, found, err := FetchKeyValueInSSTable(f.DirEntry.Name(), key)
		if found {
			return val, nil
		}
		if errors.Is(err, ErrKeyNotFound) {
			continue
		}
		if err != nil {
			return "", fmt.Errorf("search %s: %w", f.DirEntry.Name(), err)
		}
	}
	return "", ErrKeyNotFound
}
