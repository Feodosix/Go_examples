package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"sync"
)

type request struct {
	URL string `json:"url"`
}

type response struct {
	URL string `json:"url"`
	Key string `json:"key"`
}

func main() {
	port := flag.String("port", "", "port to listen on")
	flag.Parse()
	if *port == "" {
		fmt.Fprintln(os.Stderr, "port is required")
		os.Exit(1)
	}

	store := newStore()

	mux := http.NewServeMux()
	mux.HandleFunc("/shorten", store.shortenHandler)
	mux.HandleFunc("/go/", store.redirectHandler)

	err := http.ListenAndServe(":"+*port, mux)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Server error:", err)
		os.Exit(1)
	}
}

type store struct {
	mu       sync.Mutex
	urlToKey map[string]string
	keyToURL map[string]string
}

func newStore() *store {
	return &store{
		urlToKey: make(map[string]string),
		keyToURL: make(map[string]string),
	}
}

func (s *store) shortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req request
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.URL == "" {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	key, exists := s.urlToKey[req.URL]
	if !exists {
		for {
			key = generateKey(10)
			if _, used := s.keyToURL[key]; !used {
				break
			}
		}
		s.urlToKey[req.URL] = key
		s.keyToURL[key] = req.URL
	}
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	resp := response{URL: req.URL, Key: key}
	json.NewEncoder(w).Encode(resp)
}

func (s *store) redirectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	key := r.URL.Path[len("/go/"):]
	s.mu.Lock()
	dest, ok := s.keyToURL[key]
	s.mu.Unlock()
	if !ok {
		http.Error(w, "key not found", http.StatusNotFound)
		return
	}
	http.Redirect(w, r, dest, http.StatusFound)
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateKey(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
