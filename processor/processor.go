package processor

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/rodrigo-kayala/mirage-mocker/config"
)

var (
	ErrNoMatchFound = errors.New("no match found for request")
)

type Processor interface {
	Process(w http.ResponseWriter, r *http.Request)
}

// Processor structure
type processor struct {
	Parsers []parser
}

// Parser interface
type parser interface {
	ProcessRequest(w http.ResponseWriter, r *http.Request)
	GetBaseParser() baseParser
}

// Process current request and write response
func (rp *processor) Process(w http.ResponseWriter, r *http.Request) {
	requestProcess, err := rp.matchParser(r)
	log.Debug().Msgf("requestProcess: %#v", requestProcess)

	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, ErrNoMatchFound) {
			status = http.StatusNotFound
		}

		errorResponse(w, fmt.Sprintf("error processing request: %v", err), status)
		return
	}
	base := requestProcess.GetBaseParser()
	delay(base.MinDelay, base.MaxDelay)

	requestProcess.ProcessRequest(w, r)
}

func delay(min time.Duration, max time.Duration) {
	delta := int64(max - min)
	if delta <= 0 {
		return
	}
	d := rand.Int63n(delta) + int64(min)

	time.Sleep(time.Duration(d))
}

func (rp *processor) matchParser(r *http.Request) (parser, error) {
	log.Debug().Msgf("parsers: %#v", rp.Parsers)
	for _, parser := range rp.Parsers {
		bp := parser.GetBaseParser()
		if bp.ContentType != "" && !strings.Contains(r.Header.Get("Content-Type"), bp.ContentType) {
			continue
		}

		if !containsMethod(bp.Methods, r.Method) {
			continue
		}

		match, err := regexp.MatchString(bp.Pattern, r.URL.Path)
		if err != nil {
			return nil, err
		}

		if match {
			return parser, nil
		}
	}
	return nil, ErrNoMatchFound
}

// baseParser base structure
type baseParser struct {
	Pattern     string
	Methods     []string
	ContentType string
	Log         bool
	MinDelay    time.Duration
	MaxDelay    time.Duration
}

// NewFromConfig creates a new RequestProcessor from a Config struct
func NewFromConfig(c config.Config) (Processor, error) {
	var proc processor
	for _, service := range c.Services {
		base, err := createBaseParser(service.Parser)
		if err != nil {
			return nil, fmt.Errorf("error parsing base config: %w", err)
		}

		switch service.Parser.ConfigType {
		case "pass":
			passParser, err := createPassParser(base, service.Parser)
			if err != nil {
				return nil, fmt.Errorf("error while creating pass parser: %w", err)
			}
			proc.Parsers = append(proc.Parsers, passParser)
		case "mock":
			mparser := mockParser{baseParser: base}
			baseResp := baseResponse{Status: service.Parser.Response.Status}

			resp, err := parseMockResponseConfig(service.Parser.Response, baseResp)
			if err != nil {
				return nil, fmt.Errorf("error while parsing response: %w", err)
			}
			mparser.Response = resp
			proc.Parsers = append(proc.Parsers, &mparser)
		default:
			return nil, fmt.Errorf("bad value for config-type %s", service.Parser.ConfigType)
		}
	}

	return &proc, nil
}

func createBaseParser(conf config.Parser) (baseParser, error) {
	base := baseParser{
		ContentType: conf.ContentType,
		Log:         conf.Log,
		Methods:     conf.Methods,
		Pattern:     conf.Pattern,
	}

	if conf.Delay.Min != "" && conf.Delay.Max != "" {
		min, err := time.ParseDuration(conf.Delay.Min)
		if err != nil {
			return baseParser{}, fmt.Errorf("error parsing min delay: %w", err)
		}
		max, err := time.ParseDuration(conf.Delay.Max)
		if err != nil {
			return baseParser{}, fmt.Errorf("error parsing max delay: %w", err)
		}
		base.MaxDelay = max
		base.MinDelay = min
	}

	return base, nil
}

func parseMockResponseConfig(conf config.Response, base baseResponse) (response, error) {
	switch conf.BodyType {

	case "fixed":
		body := conf.Body
		if conf.BodyFile != "" {
			var err error
			body, err = readBodyFile(conf.BodyFile)
			if err != nil {
				return nil, fmt.Errorf("failed to open body response file: %w", err)
			}
		}

		return &responseFixed{
			baseResponse: base,
			Body:         body,
			ContentType:  conf.ContentType,
		}, nil
	case "request":
		return &responseRequest{
			baseResponse: base,
			ContentType:  conf.ContentType,
		}, nil
	case "runnable":
		runnable, err := loadRunnableFunc(conf.ResponseLib, conf.ResponseSymbol)
		if err != nil {
			return nil, fmt.Errorf("error processing transform method: %w", err)
		}

		return &responseRunnable{
			baseResponse: base,
			runnable:     runnable,
		}, nil
	default:
		return nil, fmt.Errorf("bad value for body-type %s", conf.BodyType)
	}
}

func readBodyFile(src string) (string, error) {
	b, err := ioutil.ReadFile(src)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func containsMethod(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
