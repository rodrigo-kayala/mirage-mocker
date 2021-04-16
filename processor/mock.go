package processor

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

// Runnable plugin structure
type runnable struct {
	runnableFunc func(w http.ResponseWriter, r *http.Request, status int) error
}

type mockParser struct {
	baseParser
	Response response
}

// ProcessRequest process mock requests
func (mr *mockParser) ProcessRequest(w http.ResponseWriter, r *http.Request) {
	if mr.Log {
		logRequest(r)
	}

	mr.Response.WriteResponse(w, r)
}

func (mr *mockParser) GetBaseParser() baseParser {
	return mr.baseParser
}

type response interface {
	WriteResponse(w http.ResponseWriter, r *http.Request)
}

type baseResponse struct {
	Status map[string]int
}

type responseFixed struct {
	baseResponse
	ContentType string
	Body        string
}

// WriteResponse writes response for fixed response type
func (rf *responseFixed) WriteResponse(w http.ResponseWriter, r *http.Request) {
	if rf.ContentType != "" {
		w.Header().Add("Content-Type", rf.ContentType)
	}

	w.WriteHeader(rf.Status[r.Method])
	w.Write([]byte(rf.Body))
}

type responseRequest struct {
	baseResponse
	ContentType string
}

// WriteResponse writes response for request response type
func (rr *responseRequest) WriteResponse(w http.ResponseWriter, r *http.Request) {
	if rr.ContentType != "" {
		w.Header().Add("Content-Type", rr.ContentType)
	}

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		errorResponse(w, fmt.Sprintf("Can't read body %v", err), 500)
		return
	}

	w.WriteHeader(rr.Status[r.Method])
	w.Write(body)
}

type responseRunnable struct {
	baseResponse
	runnable runnable
}

// WriteResponse writes response for runnable response type
func (rr *responseRunnable) WriteResponse(w http.ResponseWriter, r *http.Request) {
	err := rr.runnable.runnableFunc(w, r, rr.Status[r.Method])

	if err != nil {
		errorResponse(w, fmt.Sprintf("error running request: %v", err), 500)
		return
	}
}
