package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProbe(t *testing.T) {
	for _, status := range []int{200, 302, 503} {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(status) }))
		err := probe(s.URL)
		s.Close()
		if (err == nil) != (status == 200) {
			t.Fatalf("status %d: %v", status, err)
		}
	}
	if err := probe(":bad"); err == nil {
		t.Fatal("invalid URL accepted")
	}
}
