package processor

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"plugin"

	"github.com/rs/zerolog/log"
)

func loadRunnableFunc(lib string, symbol string) (runnable, error) {
	p, err := plugin.Open(lib)
	if err != nil {
		return runnable{}, err
	}

	s, err := p.Lookup(symbol)
	if err != nil {
		return runnable{}, err
	}

	f, ok := s.(func(w http.ResponseWriter, r *http.Request, status int) error)
	if !ok {
		return runnable{}, errors.New("runnable symbol must have this signature: func(w http.ResponseWriter, r *http.Request, status int) error")
	}

	return runnable{runnableFunc: f}, nil
}

func loadTransformFunc(lib string, symbol string) (transform, error) {
	p, err := plugin.Open(lib)
	if err != nil {
		return transform{}, err
	}

	s, err := p.Lookup(symbol)
	if err != nil {
		return transform{}, err
	}

	f, ok := s.(func(r *http.Request) error)
	if !ok {
		return transform{}, errors.New("transform symbol must have this signature: func(r *http.Request) error")
	}

	return transform{tranformFunc: f}, nil
}

func errorResponse(w http.ResponseWriter, message string, status int) {
	w.Header().Add("content-type", "text/plain")
	w.WriteHeader(status)
	log.Error().Msgf("%d %s", status, message)
	fmt.Fprint(w, message)
}

func logRequest(r *http.Request) {
	body, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Error().Err(err).Msgf("error logging request: %v", err)
		return
	}

	log.Info().Msgf("Request: %s", body)
}

func logResponse(resp *http.Response) {

	respBody, err := httputil.DumpResponse(resp, true)
	if err != nil {
		log.Error().Err(err).Msg("error logging response")
		return
	}

	log.Info().Msgf("Response: %s", respBody)
}
