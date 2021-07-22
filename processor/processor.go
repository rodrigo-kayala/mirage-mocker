package processor

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"sort"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/rodrigo-kayala/mirage-mocker/config"
)

var (
	ErrNoMatchFound  = errors.New("no match found for request")
	ErrInvalidConfig = errors.New("invalid config")
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

	requestProcess.ProcessRequest(w, r)
}

func matchHeaders(header http.Header, expectedHeaders map[string]string) bool {
	for k, v := range expectedHeaders {
		if header.Get(k) != v {
			return false
		}
	}
	return true
}

func (rp *processor) matchParser(r *http.Request) (parser, error) {
	log.Debug().Msgf("parsers: %#v", rp.Parsers)
	for _, parser := range rp.Parsers {
		bp := parser.GetBaseParser()
		if !matchHeaders(r.Header, bp.Headers) {
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
	Pattern string
	Methods []string
	Headers map[string]string
	Log     bool
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
			respCfg, err := sortDistribution(service.Parser.Responses)
			if err != nil {
				return nil, fmt.Errorf("error while parsing response distribution: %w", err)
			}
			mparser := mockParser{baseParser: base, Responses: make([]response, len(respCfg))}

			for i, response := range respCfg {
				resp, err := buildResponse(response)
				if err != nil {
					return nil, fmt.Errorf("error while building responses: %w", err)
				}

				mparser.Responses[i] = resp
			}

			proc.Parsers = append(proc.Parsers, &mparser)
		default:
			return nil, fmt.Errorf("bad value for config-type %s", service.Parser.ConfigType)
		}
	}

	return &proc, nil
}

func buildResponse(respCfg config.Response) (response, error) {
	baseResp := baseResponse{
		Status:        respCfg.Status,
		Headers:       respCfg.Headers,
		Distribuition: respCfg.Distribuition,
	}

	if respCfg.Delay.Min != "" && respCfg.Delay.Max != "" {
		min, err := time.ParseDuration(respCfg.Delay.Min)
		if err != nil {
			return nil, fmt.Errorf("error parsing min delay: %w", err)
		}
		max, err := time.ParseDuration(respCfg.Delay.Max)
		if err != nil {
			return nil, fmt.Errorf("error parsing max delay: %w", err)
		}
		baseResp.MaxDelay = max
		baseResp.MinDelay = min
	}

	return parseMockResponseConfig(respCfg, baseResp)
}

func sortDistribution(respCfg []config.Response) ([]config.Response, error) {
	if len(respCfg) == 0 {
		return nil, fmt.Errorf("empty response array: %w", ErrInvalidConfig)
	}

	outResponses := make([]config.Response, len(respCfg))
	copy(outResponses, respCfg)

	if len(outResponses) == 1 {
		outResponses[0].Distribuition = 100
		return outResponses, nil
	}

	total := 0.0
	for _, response := range outResponses {
		total += response.Distribuition
	}

	if total != 100 {
		return nil, fmt.Errorf("sum of distributions must be 100: %w", ErrInvalidConfig)
	}

	sort.Slice(outResponses, func(i, j int) bool {
		return outResponses[i].Distribuition < outResponses[j].Distribuition
	})

	return outResponses, nil
}

func createBaseParser(conf config.Parser) (baseParser, error) {
	base := baseParser{
		Headers: conf.Headers,
		Log:     conf.Log,
		Methods: conf.Methods,
		Pattern: conf.Pattern,
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
			MagicHeader: MagicHeader{
				Name:         conf.MagicHeaderName,
				SourceFolder: conf.MagicHeaderFolder,
			},
		}, nil
	case "echo":
		return &responseEcho{
			baseResponse: base,
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
