package sstable

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/huandu/skiplist"
)

type SSTEntry struct {
	KeyLength   int
	Key         string
	ValueLength int
	Value       string
}

const sstDefaultPath = "."

func FlushToSSTable(list *skiplist.SkipList) (string, error) {
	ts := time.Now().UTC().Format("2006_01_02_15:04:05")
	filename := fmt.Sprintf("sst_%v.db", ts)
	fullPath := filepath.Join(sstDefaultPath, filename)
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
