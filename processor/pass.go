package processor

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"

	"github.com/rs/zerolog/log"

	"github.com/rodrigo-kayala/mirage-mocker/config"
)

// Transform plugin structure
type transformFunc func(r *http.Request) error

// passParser type structure
type passParser struct {
	baseParser
	proxy     *httputil.ReverseProxy
	transform transformFunc
}

// ProcessRequest process pass requests
func (pp passParser) ProcessRequest(w http.ResponseWriter, r *http.Request) {
	pp.proxy.ServeHTTP(w, r)
}

// GetBaseParser returns base request
func (pp passParser) GetBaseParser() baseParser {
	return pp.baseParser
}

type logTransport struct{}

func (t *logTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := http.DefaultTransport.RoundTrip(request)
	if err != nil {
		return response, err
	}
	logResponse(request, response)

	return response, err
}

func createPassParser(base baseParser, cr config.Parser) (passParser, error) {
	var parser passParser
	parser.baseParser = base

	url, err := url.Parse(cr.PassBaseURI)
	if err != nil {
		return passParser{}, fmt.Errorf("error parsing pass url %s: %w", cr.PassBaseURI, err)
	}

	log.Debug().Msgf("PASS URL: %v", url)

	var transform transformFunc

	if cr.TransformLib != "" && cr.TransformSymbol != "" {
		transform, err = loadTransformFunc(cr.TransformLib, cr.TransformSymbol)
		if err != nil {
			return passParser{}, fmt.Errorf("error loading tranform funcion: %w", err)
		}
	}

	director := func(req *http.Request) {
		if transform != nil {
			err := transform(req)
			if err != nil {
				log.Error().Err(err).Msg("error transforming pass request")
			}
		}

		req.Host = url.Host
		req.URL.Scheme = url.Scheme
		req.URL.Host = url.Host

		for _, rw := range cr.Rewrites {

			regex, err := regexp.Compile(rw.Source)

			if err != nil {
				log.Error().Err(err).Msg("error parsing rewrite")
				continue
			}

			req.URL.Path = regex.ReplaceAllString(req.URL.Path, rw.Target)
		}

		log.Debug().Msgf("URL After rewrite: %v", req.URL)
	}

	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.Director = director
	if cr.Log {
		proxy.Transport = &logTransport{}
	}

	parser.proxy = proxy
	parser.transform = transform

	return parser, nil
}

func logResponse(req *http.Request, resp *http.Response) {

	reqBody, err := httputil.DumpRequest(req, true)
	if err != nil {
		log.Error().Err(err).Msg("error logging request)")
	}

	respBody, err := httputil.DumpResponse(resp, true)
	if err != nil {
		log.Error().Err(err).Msg("error logging response")
	}

	log.Info().Msgf("REQUEST %s\n\nRESPONSE %s", reqBody, respBody)
}
