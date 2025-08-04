package kvstore

import (
	"testing"
)

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
