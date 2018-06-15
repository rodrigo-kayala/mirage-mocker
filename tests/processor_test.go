package main

import (
	"net/http"
	"net/http/httptest"

	"testing"

	"github.com/miraeducation/mirage-mocker/config"
	"github.com/miraeducation/mirage-mocker/processor"
)

func TestProcessHandler(t *testing.T) {
	c := config.LoadConfig("config_test.yml")
	p, _ := processor.NewFromConfig(c)

	req, err := http.NewRequest("GET", "/ping", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(p.Process)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	body := "pong"
	if rr.Body.String() != body {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), body)
	}

	contentType := "text/plain"
	if rr.Header().Get("Content-Type") != contentType {
		t.Errorf("handler returned unexpected content-type: got %v want %v",
			rr.Header().Get("Content-Type"), contentType)
	}

}
