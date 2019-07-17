package sentryecho

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/lob/sentry-echo/pkg/sentry"

	"github.com/labstack/echo"
	"github.com/lob/sentry-echo/internal/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// jsonError is an Error type which can be marshalled to JSON
type jsonError struct {
	msg string
}

// Error implements the error interface
func (e jsonError) Error() string {
	return e.msg
}

// MarshalJSON implements the Marshaller interface
func (e jsonError) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.msg)
}

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

	t.Run("does not write the response if previously sent", func(t *testing.T) {
		h := handler{reporter: mockSentryClient{}}
		c, rr := test.NewContext(t, nil)
		err := errors.New("foo")

		//send a response before reporting the error
		require.NoError(t, c.String(http.StatusOK, "This is fine."))
		h.handleError(err, c)

		assert.Equal(t, http.StatusOK, rr.Code, "Status overwritten")
		assert.Equal(t, "This is fine.", rr.Body.String(), "Body overwritten")
	})

	t.Run("marshals error as JSON if supported", func(t *testing.T) {
		h := handler{reporter: mockSentryClient{}}
		c, rr := test.NewContext(t, nil)
		err := jsonError{msg: "Invalid value"}

		h.handleError(err, c)

		assert.Equal(t, http.StatusInternalServerError, rr.Code, "expected Server error status")
		assert.Equal(t, "\"Invalid value\"", rr.Body.String(), "expected marshalled body")
	})

	t.Run("marshals internal error as JSON for HTTP errors", func(t *testing.T) {
		h := handler{reporter: mockSentryClient{}}
		c, rr := test.NewContext(t, nil)
		err := echo.NewHTTPError(http.StatusBadRequest).SetInternal(
			jsonError{msg: "Invalid value"},
		)

		h.handleError(err, c)

		assert.Equal(t, http.StatusBadRequest, rr.Code, "expected status from HTTP Error")
		assert.Equal(t, "\"Invalid value\"", rr.Body.String(), "expected body from wrapped error")
	})
}

func TestHandlerIntegration(t *testing.T) {
	s, err := sentry.New("")
	require.NoError(t, err)

	e := echo.New()
	RegisterErrorHandler(e, &s)
}
