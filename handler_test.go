package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRepositoryNotFound(t *testing.T) {
	h, err := NewHandler()
	if err != nil {
		t.Fatal(err)
	}

	request, err := http.NewRequest("GET", "/notfound", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	handler := http.HandlerFunc(h.showTree)
	handler.ServeHTTP(recorder, request)

	status := recorder.Code
	if status != http.StatusNotFound {
		t.Errorf("wrong status code: got %v want %v", status, http.StatusNotFound)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "Page not found") {
		t.Errorf("body does not contains 'Page not found'")
	}
}
