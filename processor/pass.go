package processor

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"

	log "github.com/sirupsen/logrus"

	"github.com/miraeducation/mirage-mocker/config"
)

// Transform plugin structure
type Transform struct {
	TranformFunc func(r *http.Request) error
}

// PassParser type structure
type PassParser struct {
	BaseParser
	proxy     *httputil.ReverseProxy
	transform Transform
}

// ProcessRequest process pass requests
func (pp PassParser) ProcessRequest(w http.ResponseWriter, r *http.Request) {
	pp.proxy.ServeHTTP(w, r)
}

// GetBaseParser returns base request
func (pp PassParser) GetBaseParser() BaseParser {
	return pp.BaseParser
}

type logTransport struct{}

func (t *logTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := http.DefaultTransport.RoundTrip(request)

	log.Debugf("Response: %v", response)

	if err == nil {
		logResponse(request, response)
	}

	return response, err
}

func createPassParser(base BaseParser, cr config.Parser) PassParser {
	parser := PassParser{}
	parser.BaseParser = base

	url, err := url.Parse(cr.PassBaseURI)
	if err != nil {
		log.Fatalf("Invalid URI %s %v", cr.PassBaseURI, err)
	}

	log.Debugf("PASS URL: %v", url)

	var transform *Transform

	if cr.TransformLib != "" && cr.TransformSymbol != "" {
		t, err := transformMethod(cr.TransformLib, cr.TransformSymbol)
		if err != nil {
			log.Fatalf("Error processos transform method %v", err)
		}
		transform = &t
	}

	director := func(req *http.Request) {
		if transform != nil {
			err := transform.TranformFunc(req)
			if err != nil {
				log.Errorf("Error transforming pass request: %v", err)
			}
		}

		req.Host = url.Host
		req.URL.Scheme = url.Scheme
		req.URL.Host = url.Host

		for _, rw := range cr.Rewrites {

			regex, err := regexp.Compile(rw.Source)

			if err != nil {
				log.Errorf("Error parsing rewrite: %v", err)
				continue
			}

			req.URL.Path = regex.ReplaceAllString(req.URL.Path, rw.Target)
		}

		log.Debugf("URL After rewrite: %v", req.URL)
	}

	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.Director = director
	if cr.Log != "disabled" {
		proxy.Transport = &logTransport{}
	}

	parser.proxy = proxy
	parser.transform = *transform

	return parser
}

func logResponse(req *http.Request, resp *http.Response) {

	reqBody, err := httputil.DumpRequest(req, true)
	if err != nil {
		log.Errorf("Error logging request: %v", err)
	}

	respBody, err := httputil.DumpResponse(resp, true)
	if err != nil {
		log.Errorf("Error logging response: %v", err)
	}

	log.Infof("REQUEST %s\n\nRESPONSE %s", reqBody, respBody)
}
