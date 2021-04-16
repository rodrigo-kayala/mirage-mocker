package processor_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rodrigo-kayala/mirage-mocker/config"
	"github.com/rodrigo-kayala/mirage-mocker/processor"
	"github.com/stretchr/testify/assert"
)

func buildTestConfig() config.Config {
	return config.Config{
		Services: []config.Service{
			{
				Parser: config.Parser{
					Pattern:    "/mock/fixed/value.*",
					Methods:    []string{"GET"},
					ConfigType: "mock",
					Log:        true,
					Response: config.Response{
						ContentType: "text/plain",
						Status: map[string]int{
							"GET": 200,
						},
						BodyType: "fixed",
						Body:     "pong",
					},
				},
			},
			{
				Parser: config.Parser{
					Pattern:    "/mock/fixed/delay.*",
					Methods:    []string{"GET"},
					ConfigType: "mock",
					Log:        true,
					Delay: config.Delay{
						Min: "2s",
						Max: "4s",
					},
					Response: config.Response{
						ContentType: "text/plain",
						Status: map[string]int{
							"GET": 200,
						},
						BodyType: "fixed",
						Body:     "pong",
					},
				},
			},
			{
				Parser: config.Parser{
					Pattern:    "/mock/fixed/file.*",
					Methods:    []string{"GET"},
					ConfigType: "mock",
					Log:        true,
					Response: config.Response{
						ContentType: "application/json",
						Status: map[string]int{
							"GET": 200,
						},
						BodyType: "fixed",
						BodyFile: "testdata/response1.json",
					},
				},
			},
			{
				Parser: config.Parser{
					Pattern:     "/mock/request.*",
					Methods:     []string{"POST", "PUT", "DELETE"},
					ContentType: "application/json",
					ConfigType:  "mock",
					Log:         true,
					Response: config.Response{
						ContentType: "application/json",
						Status: map[string]int{
							"POST":   201,
							"PUT":    200,
							"DELETE": 204,
						},
						BodyType: "request",
					},
				},
			},
		},
	}
}

func Test_processor_Process(t *testing.T) {
	type args struct {
		config      config.Config
		method      string
		endpoint    string
		body        io.Reader
		contentType string
	}
	type out struct {
		status          int
		body            string
		contentType     string
		minimumDuration time.Duration
	}
	type test struct {
		name    string
		args    args
		out     out
		wantErr bool
	}

	tests := []test{
		{
			name: "no matches",
			args: args{
				config:   buildTestConfig(),
				method:   "GET",
				endpoint: "/other",
				body:     nil,
			},
			out: out{
				status:          404,
				body:            "error processing request: no match found for request",
				contentType:     "text/plain",
				minimumDuration: 0,
			},
			wantErr: false,
		},
		{
			name: "mock with fixed text value",
			args: args{
				config:   buildTestConfig(),
				method:   "GET",
				endpoint: "/mock/fixed/value/something",
				body:     nil,
			},
			out: out{
				status:          200,
				body:            "pong",
				contentType:     "text/plain",
				minimumDuration: 0,
			},
			wantErr: false,
		},
		{
			name: "mock with fixed text with delay",
			args: args{
				config:   buildTestConfig(),
				method:   "GET",
				endpoint: "/mock/fixed/delay/something",
				body:     nil,
			},
			out: out{
				status:          200,
				body:            "pong",
				contentType:     "text/plain",
				minimumDuration: 5 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "mock with fixed file value",
			args: args{
				config:   buildTestConfig(),
				method:   "GET",
				endpoint: "/mock/fixed/file/something",
				body:     nil,
			},
			out: out{
				status:          200,
				body:            "{\"some\": \"response\"}",
				contentType:     "application/json",
				minimumDuration: 0,
			},
			wantErr: false,
		},
		{
			name: "mock with response equals request",
			args: args{
				config:      buildTestConfig(),
				method:      "POST",
				endpoint:    "/mock/request/something",
				body:        strings.NewReader("{\"some\": \"response\"}"),
				contentType: "application/json",
			},
			out: out{
				status:          201,
				body:            "{\"some\": \"response\"}",
				contentType:     "application/json",
				minimumDuration: 0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			assert := assert.New(t)

			p, err := processor.NewFromConfig(tt.args.config)
			if tt.wantErr {
				assert.Error(err)
				return
			}
			assert.NoError(err)

			req, err := http.NewRequest(tt.args.method, tt.args.endpoint, tt.args.body)
			assert.NoError(err)

			if tt.args.contentType != "" {
				req.Header.Add("Content-Type", tt.args.contentType)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(p.Process)

			handler.ServeHTTP(rr, req)
			assert.Equal(tt.out.status, rr.Code)
			assert.Equal(tt.out.body, rr.Body.String())
			assert.Equal(tt.out.contentType, rr.Header().Get("Content-Type"))
			elapsed := time.Now().Sub(start)

			assert.LessOrEqual(tt.out.minimumDuration, elapsed)
		})
	}
}

// func TestMockFixedResponse(t *testing.T) {
// 	assert := assert.New(t)

// 	c := config.LoadConfig("../examples/config.yml")
// 	p, _ := NewFromConfig(c)

// 	req, err := http.NewRequest("GET", "/ping", nil)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	rr := httptest.NewRecorder()
// 	handler := http.HandlerFunc(p.Process)

// 	handler.ServeHTTP(rr, req)

// 	assert.EqualValues(http.StatusOK, rr.Code)
// 	assert.EqualValues("pong", rr.Body.String())
// 	assert.EqualValues("text/plain", rr.Header().Get("Content-Type"))

// }

// func TestMockRequestResponse(t *testing.T) {
// 	assert := assert.New(t)
// 	body := map[string]string{"teste": "teste1"}
// 	jsonBody, err := json.Marshal(body)

// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	c := config.LoadConfig("../examples/config.yml")
// 	p, _ := NewFromConfig(c)

// 	req, err := http.NewRequest("POST", "/teste", bytes.NewReader(jsonBody))
// 	req.Header.Add("Content-Type", "application/json")
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	rr := httptest.NewRecorder()
// 	handler := http.HandlerFunc(p.Process)

// 	handler.ServeHTTP(rr, req)
// 	respBody := make(map[string]string)
// 	err = json.Unmarshal(rr.Body.Bytes(), &respBody)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	assert.EqualValues(http.StatusCreated, rr.Code)
// 	assert.EqualValues(body, respBody)
// 	assert.EqualValues("application/json", rr.Header().Get("Content-Type"))

// }

// func TestMockRunnableResponse(t *testing.T) {
// 	assert := assert.New(t)

// 	c := config.LoadConfig("../examples/config.yml")
// 	p, _ := NewFromConfig(c)

// 	req, err := http.NewRequest("GET", "/version", nil)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	rr := httptest.NewRecorder()
// 	handler := http.HandlerFunc(p.Process)

// 	handler.ServeHTTP(rr, req)

// 	assert.EqualValues(http.StatusOK, rr.Code)
// 	assert.EqualValues("v1.0.0", rr.Body.String())
// 	assert.EqualValues("text/plain", rr.Header().Get("Content-Type"))

// }

// func TestPassResponse(t *testing.T) {
// 	assert := assert.New(t)
// 	body := map[string]string{"teste": "teste1"}
// 	jsonBody, err := json.Marshal(body)

// 	backend := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		assert.EqualValues(r.Header.Get("VERSION"), "v1.0.1")
// 		assert.EqualValues(r.URL.Path, "/pass")

// 		w.Header().Add("Content-Type", "application/json")
// 		w.WriteHeader(200)

// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		w.Write(jsonBody)
// 	}))

// 	backend.Start()
// 	defer backend.Close()

// 	c := config.LoadConfig("../examples/config.yml")
// 	c.Services[3].Parser.PassBaseURI = backend.URL

// 	p, _ := NewFromConfig(c)

// 	req, err := http.NewRequest("GET", "/test/pass", nil)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	rr := httptest.NewRecorder()
// 	handler := http.HandlerFunc(p.Process)

// 	handler.ServeHTTP(rr, req)

// 	assert.EqualValues(http.StatusOK, rr.Code)
// 	assert.EqualValues(jsonBody, rr.Body.String())
// 	assert.EqualValues("application/json", rr.Header().Get("Content-Type"))

// }

// func Test_processor_Process(t *testing.T) {
// 	type fields struct {
// 		Parsers []parser
// 	}
// 	type args struct {
// 		w http.ResponseWriter
// 		r *http.Request
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		args   args
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			rp := &processor{
// 				Parsers: tt.fields.Parsers,
// 			}
// 			rp.Process(tt.args.w, tt.args.r)
// 		})
// 	}
// }
