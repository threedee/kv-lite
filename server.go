package main

import (
	"io/ioutil"
	"net/http"
)

// Command for the channel
type Command struct {
	ty        CommandType
	key       string
	val       []byte
	replyChan chan reply
}

// CommandType to be used
type CommandType int

// Commands for the server
const (
	GetCommand = iota
	PutCommand
	DeleteCommand
)

// Server to manage the channel
type Server struct {
	cmds chan<- Command
}

type reply struct {
	val   []byte
	found bool
}

func (s *Server) handler(w http.ResponseWriter, req *http.Request) {
	key := req.URL.Path[len("/"):]

	switch req.Method {
	case http.MethodPut:
		defer req.Body.Close()
		val, err := ioutil.ReadAll(req.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			replyChan := make(chan reply)
			s.cmds <- Command{ty: PutCommand, key: key, val: val, replyChan: replyChan}
			_ = <-replyChan
			w.WriteHeader(http.StatusNoContent)
		}
	case http.MethodGet:
		replyChan := make(chan reply)
		s.cmds <- Command{ty: GetCommand, key: key, replyChan: replyChan}
		reply := <-replyChan

		if reply.found {
			w.Write(reply.val)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	case http.MethodDelete:
		replyChan := make(chan reply)
		s.cmds <- Command{ty: DeleteCommand, key: key, replyChan: replyChan}
		_ = replyChan

		w.WriteHeader(http.StatusNoContent)
	}
}
