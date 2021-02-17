package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

// KVStore uses a lock for map
type KVStore struct {
	mu       sync.RWMutex
	m        map[string][]byte
	fileName string
}

// Get locks map and retrieves value from map
func (k *KVStore) Get(w http.ResponseWriter, key string) (value []byte, found bool) {
	value, found = k.m[key]
	return
}

// Set the value for given key with locking map
func (k *KVStore) Set(w http.ResponseWriter, key string, value []byte) {
	k.m[key] = value
	k.save()
}

// Delete key from map with Lock
func (k *KVStore) Delete(w http.ResponseWriter, key string) {
	delete(k.m, key)
	k.save()
}

func (k *KVStore) save() error {
	k.mu.Lock()
	defer k.mu.Unlock()

	var err error
	if file, err := json.Marshal(k.m); err == nil {
		if err = ioutil.WriteFile(k.fileName, file, 0644); err == nil {
			return nil
		}
	}
	return fmt.Errorf("problem saving file %s, %v", k.fileName, err)

}

func (k *KVStore) init() {
	k.mu.Lock()
	defer k.mu.Unlock()

	if file, err := ioutil.ReadFile(k.fileName); err == nil {
		if err = json.Unmarshal(file, &k.m); err == nil {
			return
		}
	}

	k.m = make(map[string][]byte)
}

func main() {
	filename := flag.String("f", "key_value.db.json", "file name used for the store.")
	port := flag.Uint("p", 5000, "Port for the server to listen on")

	flag.Parse()
	store := &KVStore{}
	store.fileName = *filename

	store.init()

	mux := http.NewServeMux()
	mux.HandleFunc("/", store.handler)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), mux); err != nil {
		log.Fatalf("could not listen on port 5000 %v", err)
	}
}
