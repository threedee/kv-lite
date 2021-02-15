package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
)

// RWMap uses a lock for map
type RWMap struct {
	sync.RWMutex
	database *json.Encoder
	m        map[string][]byte
}

// Get locks map and retrieves value from map
func (r *RWMap) Get(w http.ResponseWriter, key string) {
	r.RLock()
	defer r.RUnlock()
	if value := r.m[key]; len(value) != 0 {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Length", strconv.Itoa(len(value)))
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
	r.database.Encode(r.m)
}

// Delete key from map with Lock
func (r *RWMap) Delete(w http.ResponseWriter, key string) {
	r.Lock()
	defer r.Unlock()
	delete(r.m, key)
	w.WriteHeader(http.StatusNoContent)
	r.database.Encode(r.m)
}

// type KeyValueServer struct {
// 	store RWMap
// }

func (r *RWMap) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	key := req.URL.Path[len("/"):]

	switch req.Method {
	case http.MethodPut:
		defer req.Body.Close()
		body, _ := ioutil.ReadAll(req.Body)
		r.Set(w, key, body)
	case http.MethodGet:
		r.Get(w, key)
	case http.MethodDelete:
		r.Delete(w, key)
	}

	fmt.Printf("%q", r.m)
}

type tape struct {
	file *os.File
}

func (t *tape) Write(p []byte) (n int, err error) {
	t.file.Truncate(0)
	t.file.Seek(0, 0)
	return t.file.Write(p)
}

const dbFileName = "key_value.db.json"

func main() {
	db, err := os.OpenFile(dbFileName, os.O_RDWR|os.O_CREATE, 0666)

	if err != nil {
		log.Fatalf("problem opening %s %v", dbFileName, err)
	}
	defer db.Close()
	if err = initialiseKeyValueDBFile(db); err != nil {
		log.Fatalf("problem getting file info from file %s, %v", db.Name(), err)
	}
	var m map[string][]byte
	err = json.NewDecoder(db).Decode(&m)
	if err != nil {
		log.Fatalf("problem parsing map, %v", err)
	}
	store := &RWMap{
		database: json.NewEncoder(&tape{db}),
		m:        m,
	}

	mux := http.NewServeMux()
	mux.Handle("/", store)

	if err := http.ListenAndServe(":5000", mux); err != nil {
		log.Fatalf("could not listen on port 5000 %v", err)
	}
}
func initialiseKeyValueDBFile(file *os.File) error {
	file.Seek(0, 0)

	info, err := file.Stat()

	if err != nil {
		return fmt.Errorf("problem getting file info from file %s, %v", file.Name(), err)
	}

	if info.Size() == 0 {
		file.Write([]byte("{}"))
		file.Seek(0, 0)
	}

	return nil
}
