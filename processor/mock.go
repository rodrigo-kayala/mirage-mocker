package processor

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"

	log "github.com/sirupsen/logrus"
)

// Runnable plugin structure
type Runnable struct {
	RunnableFunc func(w http.ResponseWriter, r *http.Request, status int) error
}

// MockParser base structure
type MockParser struct {
	BaseParser
	Response Response
}

// ProcessRequest process mock requests
func (mr MockParser) ProcessRequest(w http.ResponseWriter, r *http.Request) {
	if mr.Log != "disabled" {
		logRequest(r)
	}

	mr.Response.WriteResponse(w, r)
}

// GetBaseParser returns base parser
func (mr MockParser) GetBaseParser() BaseParser {
	return mr.BaseParser
}

// Response interface
type Response interface {
	WriteResponse(w http.ResponseWriter, r *http.Request)
}

// BaseResponse base structure
type BaseResponse struct {
	Status map[string]int
}

// ResponseFixed structure
type ResponseFixed struct {
	BaseResponse
	ContentType string
	Body        string
}

// WriteResponse writes response for fixed response type
func (rf ResponseFixed) WriteResponse(w http.ResponseWriter, r *http.Request) {
	if rf.ContentType != "" {
		w.Header().Add("Content-Type", rf.ContentType)
	}

	w.WriteHeader(rf.Status[r.Method])
	w.Write([]byte(rf.Body))
}

// ResponseRequest structure
type ResponseRequest struct {
	BaseResponse
	ContentType string
}

// WriteResponse writes response for request response type
func (rr ResponseRequest) WriteResponse(w http.ResponseWriter, r *http.Request) {
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

// ResponseRunnable structure
type ResponseRunnable struct {
	BaseResponse
	runnable Runnable
}

// WriteResponse writes response for runnable response type
func (rr ResponseRunnable) WriteResponse(w http.ResponseWriter, r *http.Request) {
	err := rr.runnable.RunnableFunc(w, r, rr.Status[r.Method])

	if err != nil {
		errorResponse(w, fmt.Sprintf("Error running request: %v", err), 500)
		return
	}
}

func logRequest(r *http.Request) {

	body, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Errorf("Error logging request: %v", err)
	}

	log.Infof("%s", body)
}
