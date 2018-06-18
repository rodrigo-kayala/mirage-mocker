package processor

import (
	"net/http"
	"net/http/httptest"

	"testing"

	"github.com/miraeducation/mirage-mocker/config"
	"github.com/stretchr/testify/assert"
)

func TestMockFixedResponse(t *testing.T) {
	assert := assert.New(t)

	c := config.LoadConfig("../test_configs/config_test.yml")
	p, _ := NewFromConfig(c)

	req, err := http.NewRequest("GET", "/ping", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(p.Process)

	handler.ServeHTTP(rr, req)

	assert.EqualValues(http.StatusOK, rr.Code)
	assert.EqualValues("pong", rr.Body.String())
	assert.EqualValues("text/plain", rr.Header().Get("Content-Type"))

}
