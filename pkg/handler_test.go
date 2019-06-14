package sentryecho

import (
	"errors"
	"net/http"
	"testing"

	"github.com/lob/sentry-echo/pkg/sentry"

	"github.com/labstack/echo"
	"github.com/lob/sentry-echo/internal/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockSentryClient struct{}

func (m mockSentryClient) Report(err error, req *http.Request) (string, chan error) {
	return "", nil
}

func TestHandler(t *testing.T) {
	t.Run("surfaces generic errors as internal server errors", func(tt *testing.T) {
		h := handler{reporter: mockSentryClient{}}
		c, rr := test.NewContext(t, nil)
		err := errors.New("foo")

		h.handleError(err, c)

		assert.Equal(tt, http.StatusInternalServerError, rr.Code, "expected generic errors to be 500s")
		assert.Contains(tt, rr.Body.String(), "Internal Server Error", "expected generic errors to have the correct message")
	})

	t.Run("surfaces HTTP errors transparently", func(tt *testing.T) {
		h := handler{reporter: mockSentryClient{}}
		c, rr := test.NewContext(t, nil)
		err := echo.NewHTTPError(http.StatusForbidden)

		h.handleError(err, c)

		assert.Equal(tt, http.StatusForbidden, rr.Code, "expected HTTP errors to be correct")
		assert.Contains(tt, rr.Body.String(), "Forbidden", "expected HTTP errors to have the correct message")
	})

	t.Run("obfuscates custom HTTP error messages", func(tt *testing.T) {
		h := handler{reporter: mockSentryClient{}}
		c, rr := test.NewContext(t, nil)
		err := echo.NewHTTPError(http.StatusForbidden, "this should not be shown")

		h.handleError(err, c)

		assert.Equal(tt, http.StatusForbidden, rr.Code, "expected HTTP errors to be correct")
		assert.Contains(tt, rr.Body.String(), "Forbidden", "expected HTTP errors to have the correct message")
	})

	t.Run("overwrites default HTTP status codes message when customErrorMessages is enabled", func(tt *testing.T) {
		h := handler{reporter: mockSentryClient{}, customErrorMessages: true}
		c, rr := test.NewContext(t, nil)
		err := echo.NewHTTPError(http.StatusBadRequest, "custom error message")

		h.handleError(err, c)

		assert.Equal(tt, http.StatusBadRequest, rr.Code, "expected HTTP errors to be correct")
		assert.Contains(tt, rr.Body.String(), "custom error message", "expected HTTP errors to have the overwritten message")
	})
}

func TestHandlerIntegration(t *testing.T) {
	s, err := sentry.New("")
	require.NoError(t, err)

	e := echo.New()
	RegisterErrorHandler(e, &s)
}
