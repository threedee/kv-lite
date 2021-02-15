package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
)

// RWMap uses a lock for map
type RWMap struct {
	sync.RWMutex
	m map[string][]byte
}

// Get locks map and retrieves value from map
func (r *RWMap) Get(w http.ResponseWriter, key string) {
	r.RLock()
	defer r.RUnlock()
	if value := r.m[key]; len(value) != 0 {
		w.Write(value)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

// Set the value for given key with locking map
func (r *RWMap) Set(w http.ResponseWriter, key string, value []byte) {
	r.Lock()
	defer r.Unlock()
	r.m[key] = value
	w.WriteHeader(http.StatusNoContent)
	r.save()
}

// Delete key from map with Lock
func (r *RWMap) Delete(w http.ResponseWriter, key string) {
	r.Lock()
	defer r.Unlock()
	delete(r.m, key)
	w.WriteHeader(http.StatusNoContent)
	r.save()
}

func (r *RWMap) save() {
	db, err := os.OpenFile(dbFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)

	if err != nil {
		log.Fatalf("problem opening %s %v", dbFileName, err)
	}
	defer db.Close()
	json.NewEncoder(db).Encode(r.m)
}

func (r *RWMap) init() {
	db, err := os.OpenFile(dbFileName, os.O_RDWR|os.O_CREATE, 0666)

	if err != nil {
		log.Fatalf("problem opening %s %v", dbFileName, err)
	}
	defer db.Close()

	info, err := db.Stat()

	if err != nil {
		log.Fatalf("problem getting file info from file %s, %v", db.Name(), err)
	}

	if info.Size() != 0 {
		if err = json.NewDecoder(db).Decode(&r.m); err != nil {
			log.Fatalf("problem parsing map, %v", err)
		}
	} else {
		r.m = make(map[string][]byte)
	}
}

func (r *RWMap) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	key := req.URL.Path[len("/"):]

	switch req.Method {
	case http.MethodPut:
		defer req.Body.Close()
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Fatalf("error reading body, %v", err)
		}
		r.Set(w, key, body)
	case http.MethodGet:
		r.Get(w, key)
	case http.MethodDelete:
		r.Delete(w, key)
	}
}

const dbFileName = "key_value.db.json"

func main() {
	store := &RWMap{}
	store.init()

	mux := http.NewServeMux()
	mux.Handle("/", store)

	if err := http.ListenAndServe(":5000", mux); err != nil {
		log.Fatalf("could not listen on port 5000 %v", err)
	}
}
