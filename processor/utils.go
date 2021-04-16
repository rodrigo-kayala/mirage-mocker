package processor

import (
	"errors"
	"fmt"
	"net/http"
	"plugin"

	"github.com/rs/zerolog/log"
)

func loadRunnableFunc(lib string, symbol string) (runnableFunc, error) {
	p, err := plugin.Open(lib)
	if err != nil {
		return nil, err
	}

	s, err := p.Lookup(symbol)
	if err != nil {
		return nil, err
	}

	f, ok := s.(runnableFunc)
	if !ok {
		return nil, errors.New("runnable symbol must have this signature: func(w http.ResponseWriter, r *http.Request) error")
	}

	return f, nil
}

func loadTransformFunc(lib string, symbol string) (transformFunc, error) {
	p, err := plugin.Open(lib)
	if err != nil {
		return nil, err
	}

	s, err := p.Lookup(symbol)
	if err != nil {
		return nil, err
	}

	f, ok := s.(transformFunc)
	if !ok {
		return nil, errors.New("transform symbol must have this signature: func(r *http.Request) error")
	}

	return f, nil
}

func errorResponse(w http.ResponseWriter, message string, status int) {
	w.Header().Add("content-type", "text/plain")
	w.WriteHeader(status)
	log.Error().Msgf("%d %s", status, message)
	fmt.Fprintf(w, message)
}
