package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"fudge/config"
)

func TestRepositoryNotFound(t *testing.T) {
	cfg := &config.Config{
		Root:  "../",
		Debug: true,
	}

	h, err := NewHandler(cfg)
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
		t.Error("body does not contains 'Page not found'")
	}
}
