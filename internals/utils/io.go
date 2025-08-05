package utils

import (
	"bufio"
	"encoding/json"
	"lsm-kv/internals/wal"
	"os"
	"path/filepath"
)

func ReadWAL(path string) ([]wal.LogEntry, error) {
	var entries []wal.LogEntry
	fullPath := filepath.Join(path, wal.WalFilename)
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var entry wal.LogEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}
