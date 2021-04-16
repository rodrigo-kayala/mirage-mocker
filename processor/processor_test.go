package processor_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
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
					Pattern:     "/.*",
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
						Min: "200ms",
						Max: "300ms",
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
					Pattern:    "/mock/runnable.*",
					Methods:    []string{"GET"},
					ConfigType: "mock",
					Log:        true,
					Response: config.Response{
						Status: map[string]int{
							"GET": 200,
						},
						BodyType:       "runnable",
						ResponseLib:    "testdata/runnable/runnable.so",
						ResponseSymbol: "GetEnv",
					},
				},
			},
		},
	}
}

func Test_processor_Process(t *testing.T) {
	os.Setenv("MIRAGE_MOCKER_TEST_VAR", "mirage-mocker")

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
				minimumDuration: 200 * time.Millisecond,
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
		{
			name: "mock with runnable response",
			args: args{
				config:   buildTestConfig(),
				method:   "GET",
				endpoint: "/mock/runnable?vname=MIRAGE_MOCKER_TEST_VAR",
			},
			out: out{
				status:          200,
				body:            "mirage-mocker",
				contentType:     "text/plain",
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

func Test_processor_Process__pass(t *testing.T) {
	assert := assert.New(t)
	os.Setenv("MIRAGE_MOCKER_TEST_VAR", "mirage-mocker")

	inBody := map[string]string{"some": "value"}
	inJson, err := json.Marshal(inBody)
	assert.NoError(err)

	outBody := map[string]string{"other": "value"}
	outJson, err := json.Marshal(outBody)
	assert.NoError(err)

	backend := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(r.Header.Get("OTHER_HEADER"), "otherHeader")
		assert.Equal(r.Header.Get("MIRAGE_MOCKER_TEST_VAR"), "mirage-mocker")
		assert.Equal(r.URL.Path, "/pass")

		body := r.Body
		defer body.Close()
		bodyBytes, err := io.ReadAll(body)
		assert.EqualValues(inJson, bodyBytes)

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(200)

		if err != nil {
			assert.NoError(err)
		}
		w.Write(outJson)
	}))

	backend.Start()
	defer backend.Close()

	c := config.Config{
		Services: []config.Service{
			{
				Parser: config.Parser{
					Pattern: "/test/pass.*",
					Rewrites: []config.Rewrite{
						{
							Source: "/test(/.*)",
							Target: "$1",
						},
					},
					Methods:         []string{"POST"},
					ConfigType:      "pass",
					Log:             true,
					TransformLib:    "testdata/transform/transform.so",
					TransformSymbol: "AddHeader",
					PassBaseURI:     backend.URL,
				},
			},
		},
	}
	p, err := processor.NewFromConfig(c)
	assert.NoError(err)

	req, err := http.NewRequest("POST", "/test/pass?vname=MIRAGE_MOCKER_TEST_VAR", bytes.NewReader(inJson))
	assert.NoError(err)
	req.Header.Add("OTHER_HEADER", "otherHeader")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(p.Process)

	handler.ServeHTTP(rr, req)

	assert.Equal(http.StatusOK, rr.Code)
	assert.Equal(string(outJson), rr.Body.String())
	assert.Equal("application/json", rr.Header().Get("Content-Type"))

}
