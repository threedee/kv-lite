package main

import (
	"io/ioutil"
	"log"
	"net/http"
)

func (k *KVStore) handler(w http.ResponseWriter, req *http.Request) {
	key := req.URL.Path[len("/"):]

	switch req.Method {
	case http.MethodPut:
		defer req.Body.Close()
		value, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Fatalf("error reading body, %v", err)
		}
		k.Set(w, key, value)

		w.WriteHeader(http.StatusNoContent)
	case http.MethodGet:
		value, found := k.Get(w, key)
		if found {
			w.Write(value)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	case http.MethodDelete:
		k.Delete(w, key)
		w.WriteHeader(http.StatusNoContent)
	}
}
