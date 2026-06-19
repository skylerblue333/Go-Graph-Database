package main

import (
"net/http"
"net/http/httptest"
"testing"
)

func TestHealth(t *testing.T) {
req, err := http.NewRequest("GET", "/health", nil)
if err != nil {
t.Fatal(err)
}
rr := httptest.NewRecorder()
healthHandler(rr, req)
if status := rr.Code; status != http.StatusOK {
t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
}
}

func TestGraph(t *testing.T) {
// Reset state
mu.Lock()
nodes = make(map[string][]string)
mu.Unlock()

req, _ := http.NewRequest("GET", "/edge?from=A&to=B", nil)
rr := httptest.NewRecorder()
addEdgeHandler(rr, req)
if rr.Code != http.StatusOK {
t.Errorf("Expected 200, got %d", rr.Code)
}
}
