package sstable

import (
	"log"
	"os"
	"testing"

	"github.com/huandu/skiplist"
)

func TestFlushToSST(t *testing.T) {
	newList := skiplist.New(skiplist.String)
	newList.Set("foo", "bar")
	filename, err := FlushToSSTable(newList)
	if err != nil {
		t.Fatalf("Failed flush SST Table: %v", err)
	}
	t.Logf("Created SST Tablew with filename: %v", filename)
	err = os.Remove(filename)
	if err != nil {
		// handle the error, e.g. file does not exist or permission denied
		log.Printf("Failed to delete file: %v", err)
	}
}

func TestFindKeyInSSTable(t *testing.T) {
	newList := skiplist.New(skiplist.String)
	newList.Set("foo", "bar")
	filename, err := FlushToSSTable(newList)
	if err != nil {
		t.Fatalf("Failed flush SST Table: %v", err)
	}
	t.Logf("Created SST Tablew with filename: %v", filename)
	value, found, err := FetchKeyValueInSSTable(filename, "foo")
	if !found || err != nil {
		t.Fatalf("Failed to find key: %v", err)
	}
	if value != "bar" {
		t.Fatalf("Wrong value found: %v", value)
	}
}
