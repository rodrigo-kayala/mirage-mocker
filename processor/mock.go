package processor

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"sync"
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
	Status  map[string]int
	Headers map[string]string
}

func (br *baseResponse) addHeaders(w http.ResponseWriter) {
	for k, v := range br.Headers {
		w.Header().Add(k, v)
	}
}

type responseFixed struct {
	baseResponse
	Body        string
	MagicHeader MagicHeader
}

type MagicHeader struct {
	Name         string
	SourceFolder string
	cache        map[string]string
	mu           sync.Mutex
}

func (rf *responseFixed) magicHeaderBody(bodyFile string) (string, error) {
	// for security reasons, only the file name is considered, to prevent unauthorized access like
	//  "../someotherfile" or "/var/somefile"
	_, file := path.Split(bodyFile)

	rf.MagicHeader.mu.Lock()
	defer rf.MagicHeader.mu.Unlock()

	if rf.MagicHeader.cache == nil {
		rf.MagicHeader.cache = make(map[string]string)
	}

	body, ok := rf.MagicHeader.cache[file]
	if !ok {
		var err error
		body, err = readBodyFile(path.Join(rf.MagicHeader.SourceFolder, file))
		if err != nil {
			return "", err
		}
		rf.MagicHeader.cache[file] = body
	}

	return body, nil
}

// WriteResponse writes response for fixed response type
func (rf *responseFixed) WriteResponse(w http.ResponseWriter, r *http.Request) {
	body := rf.Body

	// if is a magic header request and the header is present
	if rf.MagicHeader.Name != "" && r.Header.Get(rf.MagicHeader.Name) != "" {
		f := r.Header.Get(rf.MagicHeader.Name)
		out, err := rf.magicHeaderBody(f)
		if err != nil {
			errorResponse(w, fmt.Sprintf("file not found %s", f), 404)
			return
		}
		body = out
	}

	rf.baseResponse.addHeaders(w)
	w.WriteHeader(rf.Status[r.Method])
	_, _ = w.Write([]byte(body))
}

type responseEcho struct {
	baseResponse
}

// WriteResponse writes response for echo response type
func (rr *responseEcho) WriteResponse(w http.ResponseWriter, r *http.Request) {
	rr.baseResponse.addHeaders(w)
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		errorResponse(w, fmt.Sprintf("Can't read body %v", err), 500)
		return
	}

	w.WriteHeader(rr.Status[r.Method])
	_, _ = w.Write(body)
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
