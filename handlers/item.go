package handlers

import (
	"encoding/json"
	"fmt"
	kv "lsm-kv/internals/kvstore"
	"net/http"
)

var Store *kv.KVStore

type Item struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func CreateItem(w http.ResponseWriter, req *http.Request) {
	var it Item
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&it)
	if err != nil {
		fmt.Printf("Error decoding JSON: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	Store.Set(it.Key, it.Value)
	w.WriteHeader(http.StatusCreated)
}

func UpdateItem(w http.ResponseWriter, req *http.Request) {
	var it Item
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&it)
	if err != nil {
		fmt.Printf("Error decoding JSON: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	Store.Set(it.Key, it.Value)
	w.WriteHeader(http.StatusOK)
}

func DeleteItem(w http.ResponseWriter, req *http.Request) {
	item := req.PathValue("id")
	value, err := Store.Get(item)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	Store.Delete(item)
	json.NewEncoder(w).Encode(value)
}

func GetItem(w http.ResponseWriter, req *http.Request) {
	item := req.PathValue("id")
	value, err := Store.Get(item)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(value)
}
