package processor

import (
	"fmt"
	"net/http"
	"plugin"

	log "github.com/sirupsen/logrus"
)

func runnableMethod(lib string, symbol string) (Runnable, error) {
	p, err := plugin.Open(lib)
	if err != nil {
		return Runnable{}, err
	}

	s, err := p.Lookup(symbol)
	if err != nil {
		return Runnable{}, err
	}

	f, ok := s.(func(w http.ResponseWriter, r *http.Request, status int) error)
	if !ok {
		return Runnable{}, fmt.Errorf("Runnable symbol must have this signature: func(w http.ResponseWriter, r *http.Request) error")
	}

	return Runnable{RunnableFunc: f}, nil
}

func transformMethod(lib string, symbol string) (Transform, error) {
	p, err := plugin.Open(lib)
	if err != nil {
		return Transform{}, err
	}

	s, err := p.Lookup(symbol)
	if err != nil {
		return Transform{}, err
	}

	f, ok := s.(func(r *http.Request) error)
	if !ok {
		return Transform{}, fmt.Errorf("Transform symbol must have this signature: func(r *http.Request) error")
	}

	return Transform{TranformFunc: f}, nil
}

func errorResponse(w http.ResponseWriter, message string, status int) {
	w.Header().Add("content-type", "text/plain")
	w.WriteHeader(status)
	log.Errorf("%s %d", message, status)
	fmt.Fprintf(w, message)
}
