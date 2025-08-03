package main

import (
	"log"
	"net/http"
	"raft-kv/handlers"
)

func main() {
	mux := http.NewServeMux()
	http.HandleFunc("GET /store", handlers.GetItem)
	http.HandleFunc("POST /store", handlers.CreateItem)
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
