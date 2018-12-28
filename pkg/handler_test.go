package sentryecho

import (
	"errors"
	"net/http"
	"testing"

	"github.com/labstack/echo"
	"github.com/lob/sentry-echo/internal/test"
	"github.com/stretchr/testify/assert"
)

type mockSentryClient struct{}

func (m mockSentryClient) Report(err error, req *http.Request) {
}

func TestHandler(t *testing.T) {
	h := handler{reporter: mockSentryClient{}}

	t.Run("surfaces generic errors as internal server errors", func(tt *testing.T) {
		c, rr := test.NewContext(t, nil)
		err := errors.New("foo")

		h.handleError(err, c)

		assert.Equal(tt, http.StatusInternalServerError, rr.Code, "expected generic errors to be 500s")
		assert.Contains(tt, rr.Body.String(), "Internal Server Error", "expected generic errors to have the correct message")
	})

	t.Run("surfaces HTTP errors transparently but obfuscates message", func(tt *testing.T) {
		c, rr := test.NewContext(t, nil)
		err := echo.NewHTTPError(http.StatusForbidden, "foo")

		h.handleError(err, c)

		assert.Equal(tt, http.StatusForbidden, rr.Code, "expected HTTP errors to be correct")
		assert.Contains(tt, rr.Body.String(), "Forbidden", "expected HTTP errors to have the correct message")
	})

	t.Run("overwrites HTTP 400 error messages", func(tt *testing.T) {
		c, rr := test.NewContext(t, nil)
		err := echo.NewHTTPError(http.StatusBadRequest, "this shouldn't be sent to customers")

		h.handleError(err, c)

		assert.Equal(tt, http.StatusBadRequest, rr.Code, "expected HTTP errors to be correct")
		assert.Contains(tt, rr.Body.String(), "Bad Request", "expected HTTP errors to have the correct message")
	})

	t.Run("overwrites HTTP 403 error messages", func(tt *testing.T) {
		c, rr := test.NewContext(t, nil)
		err := echo.NewHTTPError(http.StatusForbidden, "this shouldn't be sent to customers")

		h.handleError(err, c)

		assert.Equal(tt, http.StatusForbidden, rr.Code, "expected HTTP errors to be correct")
		assert.Contains(tt, rr.Body.String(), "Forbidden", "expected HTTP errors to have the correct message")
	})

	t.Run("overwrites HTTP 404 error messages", func(tt *testing.T) {
		c, rr := test.NewContext(t, nil)
		err := echo.NewHTTPError(http.StatusNotFound, "this shouldn't be sent to customers")

		h.handleError(err, c)

		assert.Equal(tt, http.StatusNotFound, rr.Code, "expected HTTP errors to be correct")
		assert.Contains(tt, rr.Body.String(), "Not Found", "expected HTTP errors to have the correct message")
	})
}
