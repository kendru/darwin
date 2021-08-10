package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

func NewHTTPServer(addr string) *http.Server {
	httpsrv := newHTTPServer()
	r := mux.NewRouter()
	r.HandleFunc("/api/{key}", httpsrv.handleKey).Methods(http.MethodGet, http.MethodPut)
	r.HandleFunc("/api/{key}/watch", httpsrv.handleWatch).Methods(http.MethodGet)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/static")))

	return &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 120 * time.Second,
	}
}

func newHTTPServer() *httpServer {
	return &httpServer{
		storage: make(map[string]string),
		clients: make(map[string]chan string),
	}
}

type httpServer struct {
	mu      sync.Mutex
	storage map[string]string

	cMu     sync.Mutex
	clients map[string]chan string
}

type Entry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (s *httpServer) handleList(w http.ResponseWriter, r *http.Request) {
	entries := make([]Entry, 0)
	for key, value := range s.storage {
		entries = append(entries, Entry{
			Key:   key,
			Value: value,
		})
	}

	err := json.NewEncoder(w).Encode(entries)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *httpServer) handleKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	s.mu.Lock()
	key, ok := vars["key"]
	s.mu.Unlock()
	if !ok {
		http.Error(w, "No key specified", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodGet {
		s.handleGet(key, w, r)
		return
	}
	if r.Method == http.MethodPut {
		s.handlePut(key, w, r)
		return
	}
}

func (s *httpServer) handleWatch(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keepalive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	vars := mux.Vars(r)
	s.mu.Lock()
	key, ok := vars["key"]
	s.mu.Unlock()
	if !ok {
		http.Error(w, "No key specified", http.StatusBadRequest)
		return
	}

	ch := make(chan string, 1)
	s.cMu.Lock()
	s.clients[key] = ch
	s.cMu.Unlock()
	defer func() {
		s.cMu.Lock()
		delete(s.clients, key)
		s.cMu.Unlock()
		defer close(ch)
	}()

	notify := w.(http.CloseNotifier).CloseNotify()
outer:
	for {
		select {
		case val := <-ch:
			fmt.Fprint(w, "event: value\n")
			fmt.Fprintf(w, "data: %s\n\n", string(val))
			flusher.Flush()
		case <-notify:
			log.Println("The client closed the connection")
			break outer
		}
	}
}

func (s *httpServer) handleGet(key string, w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	val, ok := s.storage[key]
	s.mu.Unlock()
	if !ok {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	}

	w.Write([]byte(val))
}

func (s *httpServer) handlePut(key string, w http.ResponseWriter, r *http.Request) {
	value, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Cannot read request body", http.StatusBadRequest)
		return
	}
	if len(value) == 0 {
		http.Error(w, "Cannot set empty value", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	s.storage[key] = string(value)
	s.mu.Unlock()

	s.cMu.Lock()
	if ch, ok := s.clients[key]; ok {
		ch <- string(value)
	}
	s.cMu.Unlock()
}
