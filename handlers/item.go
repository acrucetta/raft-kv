package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

var store = make(map[string]any)

type Item struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
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
	store[it.Key] = it.Value
	w.WriteHeader(http.StatusCreated)
}

func GetItem(w http.ResponseWriter, req *http.Request) {
	item := req.PathValue("id")
	value, ok := store[item]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(value)
}
