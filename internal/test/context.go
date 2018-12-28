package test

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
)

// NewContext returns a new echo.Context, and *httptest.ResponseRecorder to be
// used for tests.
func NewContext(t *testing.T, payload []byte) (echo.Context, *httptest.ResponseRecorder) {
	t.Helper()

	e := echo.New()
	req := httptest.NewRequest(echo.GET, "/", bytes.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rr := httptest.NewRecorder()
	c := e.NewContext(req, rr)
	return c, rr
}
