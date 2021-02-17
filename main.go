package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	filename := flag.String("f", "key_value.db.json", "file name used for the store.")
	port := flag.Uint("p", 5000, "Port for the server to listen on")

	flag.Parse()

	server := Server{StartStoreManager(filename)}

	mux := http.NewServeMux()
	mux.HandleFunc("/", server.handler)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), mux); err != nil {
		log.Fatalf("could not listen on port 5000 %v", err)
	}
}
