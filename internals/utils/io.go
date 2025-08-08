package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"lsm-kv/internals/wal"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Extracts the timestamp from the sst files. E.g., sst_2025_08_07_16:28:10.db
func ParseTimestamp(name string) (int64, error) {
	// Expected format: sst_YYYY_MM_DD_HH:MM:SS.db
	parts := strings.Split(name, "_")
	if len(parts) < 5 {
		return 0, fmt.Errorf("invalid sst filename format: %s", name)
	}
	date := fmt.Sprintf("%s-%s-%s", parts[1], parts[2], parts[3])
	timePart := strings.TrimSuffix(parts[4], ".db")
	datetime := fmt.Sprintf("%s %s", date, timePart)
	// Parse using Go's time package
	t, err := ParseSSTTimestamp(datetime)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}

func ParseSSTTimestamp(datetime string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", datetime)
}

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
