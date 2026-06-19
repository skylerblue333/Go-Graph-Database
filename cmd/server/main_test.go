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
	req, _ := http.NewRequest("GET", "/edge?from=A&to=B", nil)
	rr := httptest.NewRecorder()
	addEdgeHandler(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rr.Code)
	}

	req2, _ := http.NewRequest("GET", "/edges?node=A", nil)
	rr2 := httptest.NewRecorder()
	getEdgesHandler(rr2, req2)
	if rr2.Body.String() != "["B"]
" {
		t.Errorf("Expected ["B"], got %s", rr2.Body.String())
	}
}

