package main

import (
"encoding/json"
"fmt"
"log"
"net/http"
"sync"
)

var (
nodes = make(map[string][]string)
mu    sync.RWMutex
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(map[string]string{"status": "healthy", "service": "Go-Graph-Database"})
}

func addEdgeHandler(w http.ResponseWriter, r *http.Request) {
from := r.URL.Query().Get("from")
to := r.URL.Query().Get("to")
if from == "" || to == "" {
http.Error(w, "Missing params", http.StatusBadRequest)
return
}
mu.Lock()
nodes[from] = append(nodes[from], to)
mu.Unlock()
w.WriteHeader(http.StatusOK)
}

func getEdgesHandler(w http.ResponseWriter, r *http.Request) {
node := r.URL.Query().Get("node")
mu.RLock()
edges := nodes[node]
mu.RUnlock()
json.NewEncoder(w).Encode(edges)
}

func main() {
http.HandleFunc("/health", healthHandler)
http.HandleFunc("/edge", addEdgeHandler)
http.HandleFunc("/edges", getEdgesHandler)
fmt.Println("Starting Go-Graph-Database on :8080")
log.Fatal(http.ListenAndServe(":8080", nil))
}
