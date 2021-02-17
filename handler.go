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
		k.Set(key, value)

		w.WriteHeader(http.StatusNoContent)
	case http.MethodGet:
		value, found := k.Get(key)
		if found {
			w.Write(value)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	case http.MethodDelete:
		k.Delete(key)
		w.WriteHeader(http.StatusNoContent)
	}
}
