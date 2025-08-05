package wal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const WalFilename = "wal.log"

type LogEntry struct {
	Command string
	Key     string
	Value   string
}

type Wal struct {
	File *os.File
	mu   sync.Mutex
}

func NewLogFile(path string) (*Wal, error) {
	if path == "" {
		path = "."
	}
	fullPath := filepath.Join(path, WalFilename)
	file, err := os.OpenFile(fullPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return &Wal{File: file, mu: sync.Mutex{}}, nil
}

func (wal *Wal) AppendEntry(entry LogEntry) error {
	wal.mu.Lock()
	defer wal.mu.Unlock()

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	_, err = wal.File.Write(append(data, '\n'))
	if err != nil {
		return err
	}
	err = wal.File.Sync()
	if err != nil {
		return err
	}
	return nil
}
