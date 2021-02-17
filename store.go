package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"sync"
)

// KVStore uses a lock for map
type KVStore struct {
	mu       sync.RWMutex
	m        map[string][]byte
	fileName string
}

// Get locks map and retrieves value from map
func (k *KVStore) get(key string) (value []byte, found bool) {
	value, found = k.m[key]
	return
}

// Set the value for given key with locking map
func (k *KVStore) set(key string, value []byte) {
	k.m[key] = value
	k.save()
}

// Delete key from map with Lock
func (k *KVStore) delete(key string) {
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

// StartStoreManager starts a goroutine that serves as a manager for
// our counters datastore. Returns a channel that's used
//to send commands to the manager
func StartStoreManager(filename *string) chan<- Command {
	store := &KVStore{fileName: *filename}

	store.init()

	cmds := make(chan Command)

	go func() {
		for cmd := range cmds {
			switch cmd.ty {
			case GetCommand:
				val, found := store.get(cmd.key)
				cmd.replyChan <- reply{val, found}
			case PutCommand:
				store.set(cmd.key, cmd.val)
				store.save()
				cmd.replyChan <- reply{}
			case DeleteCommand:
				store.delete(cmd.key)
				store.save()
				cmd.replyChan <- reply{}
			default:
				log.Fatal("unknown command type", cmd.ty)

			}
		}
	}()
	return cmds
}
