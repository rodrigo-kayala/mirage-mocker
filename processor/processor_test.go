package processor

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"testing"

	"github.com/miraeducation/mirage-mocker/config"
	"github.com/stretchr/testify/assert"
)

func TestMockNotMatch(t *testing.T) {
	assert := assert.New(t)

	c := config.LoadConfig("../examples/config.yml")
	p, _ := NewFromConfig(c)

	req, err := http.NewRequest("GET", "/teste", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(p.Process)

	handler.ServeHTTP(rr, req)

	assert.EqualValues(http.StatusInternalServerError, rr.Code)
	assert.EqualValues("Error processing request: No match found for request", rr.Body.String())
	assert.EqualValues("text/plain", rr.Header().Get("Content-Type"))

}

func TestMockFixedResponse(t *testing.T) {
	assert := assert.New(t)

	c := config.LoadConfig("../examples/config.yml")
	p, _ := NewFromConfig(c)

	req, err := http.NewRequest("GET", "/ping", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(p.Process)

	handler.ServeHTTP(rr, req)

	assert.EqualValues(http.StatusOK, rr.Code)
	assert.EqualValues("pong", rr.Body.String())
	assert.EqualValues("text/plain", rr.Header().Get("Content-Type"))

}

func TestMockRequestResponse(t *testing.T) {
	assert := assert.New(t)
	body := map[string]string{"teste": "teste1"}
	jsonBody, err := json.Marshal(body)

	if err != nil {
		t.Fatal(err)
	}

	c := config.LoadConfig("../examples/config.yml")
	p, _ := NewFromConfig(c)

	req, err := http.NewRequest("POST", "/teste", bytes.NewReader(jsonBody))
	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(p.Process)

	handler.ServeHTTP(rr, req)
	respBody := make(map[string]string)
	err = json.Unmarshal(rr.Body.Bytes(), &respBody)
	if err != nil {
		t.Fatal(err)
	}

	assert.EqualValues(http.StatusCreated, rr.Code)
	assert.EqualValues(body, respBody)
	assert.EqualValues("application/json", rr.Header().Get("Content-Type"))

}

func TestMockRunnableResponse(t *testing.T) {
	assert := assert.New(t)

	c := config.LoadConfig("../examples/config.yml")
	p, _ := NewFromConfig(c)

	req, err := http.NewRequest("GET", "/version", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(p.Process)

	handler.ServeHTTP(rr, req)

	assert.EqualValues(http.StatusOK, rr.Code)
	assert.EqualValues("v1.0.0", rr.Body.String())
	assert.EqualValues("text/plain", rr.Header().Get("Content-Type"))

}

func TestPassResponse(t *testing.T) {
	assert := assert.New(t)
	body := map[string]string{"teste": "teste1"}
	jsonBody, err := json.Marshal(body)

	backend := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.EqualValues(r.Header.Get("VERSION"), "v1.0.1")
		assert.EqualValues(r.URL.Path, "/pass")

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(200)

		if err != nil {
			t.Fatal(err)
		}
		w.Write(jsonBody)
	}))

	backend.Start()
	defer backend.Close()

	c := config.LoadConfig("../examples/config.yml")
	c.Services[3].Parser.PassBaseURI = backend.URL

	p, _ := NewFromConfig(c)

	req, err := http.NewRequest("GET", "/test/pass", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(p.Process)

	handler.ServeHTTP(rr, req)

	assert.EqualValues(http.StatusOK, rr.Code)
	assert.EqualValues(jsonBody, rr.Body.String())
	assert.EqualValues("application/json", rr.Header().Get("Content-Type"))

}
