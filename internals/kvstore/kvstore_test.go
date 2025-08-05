package kvstore

import (
	"lsm-kv/internals/wal"
	"testing"
)

func TestRestoreFromWAL(t *testing.T) {

	// Use a temp dir for isolation
	tmpDir := t.TempDir()

	// Create a WAL and write some entries
	w, err := wal.NewLogFile(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create WAL: %v", err)
	}
	defer w.File.Close()

	// Write a SET and a DELETE
	err = w.AppendEntry(wal.LogEntry{Command: "SET", Key: "foo", Value: "bar"})
	if err != nil {
		t.Fatalf("Failed to append SET: %v", err)
	}
	err = w.AppendEntry(wal.LogEntry{Command: "DELETE", Key: "foo", Value: ""})
	if err != nil {
		t.Fatalf("Failed to append DELETE: %v", err)
	}

	// Now create a new KVStore, which should replay the WAL
	kv := NewKVStore(tmpDir)

	// "foo" should not exist after replay
	_, err = kv.Get("foo")
	if err == nil {
		t.Errorf("Expected error when getting deleted key after WAL replay, got value")
	}
}

func TestSetReturnsValue(t *testing.T) {
	kv := NewKVStore(".") // Initialize kv as a new key-value store instance
	err := kv.Set("foo", "bar")
	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	val, err := kv.Get("foo")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if val != "bar" {
		t.Errorf("Expected value 'bar', got '%s'", val)
	}
}

func TestDeleteRemovesValue(t *testing.T) {
	kv := NewKVStore(".") // Initialize kv as a new key-value store instance
	err := kv.Set("foo", "bar")
	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	err = kv.Delete("foo")
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	value, err := kv.Get("foo")
	if err == nil {
		t.Errorf("Expected error when getting deleted key, got value: '%s'", value)
	}
}
