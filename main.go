package main

import (
	"log"
	"lsm-kv/handlers"
	"lsm-kv/internals/kvstore"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	handlers.Store = kvstore.NewKVStore()
	http.HandleFunc("GET /store", handlers.GetItem)
	http.HandleFunc("POST /store", handlers.CreateItem)
	http.HandleFunc("DELETE /store", handlers.DeleteItem)
	http.HandleFunc("PUT /store", handlers.UpdateItem)
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
