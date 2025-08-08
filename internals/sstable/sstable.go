package sstable

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
		val := []byte(elem.Key().(string))

		// Format: [key len][key][val len][val]
		binary.Write(file, binary.LittleEndian, int32(len(key)))
		file.Write(key)
		binary.Write(file, binary.LittleEndian, int32(len(val)))
		file.Write(val)
	}
	return filename, nil
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

func FindKeyInSSTable(entry os.DirEntry, searchKey string) (string, bool, error) {
	// TODO: To implement.
	// We have a Byte file, we need to open it, then read
	// the contents of it to find the key/value pair
	filePath := filepath.Join(SstDefaultPath, entry.Name())
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
