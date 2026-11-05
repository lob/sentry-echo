package sentryecho

import (
	"errors"
	"net/http"
	"testing"

	"github.com/lob/sentry-echo/v2/pkg/sentry"

	"github.com/labstack/echo/v5"
	"github.com/lob/sentry-echo/v2/internal/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockSentryClient struct{}

func (m mockSentryClient) Report(err error, req *http.Request) (string, chan error) {
	return "", nil
}

type recordingReporter struct {
	reported []error
}

func (r *recordingReporter) Report(err error, req *http.Request) (string, chan error) {
	r.reported = append(r.reported, err)
	return "", nil
}

func TestHandlerReporting(t *testing.T) {
	t.Run("reports internal server errors to Sentry", func(tt *testing.T) {
		reporter := &recordingReporter{}
		h := handler{reporter: reporter}
		c, rr := test.NewContext(tt, nil)

		h.handleError(c, errors.New("foo"))

		assert.Equal(tt, http.StatusInternalServerError, rr.Code)
		assert.Len(tt, reporter.reported, 1)
	})

	t.Run("keeps the status of echo's own errors and does not report them", func(tt *testing.T) {
		reporter := &recordingReporter{}
		h := handler{reporter: reporter}
		c, rr := test.NewContext(tt, nil)

		h.handleError(c, echo.ErrNotFound)

		assert.Equal(tt, http.StatusNotFound, rr.Code)
		assert.Contains(tt, rr.Body.String(), "Not Found")
		assert.Empty(tt, reporter.reported)
	})

	t.Run("keeps the status of a wrapped HTTP error", func(tt *testing.T) {
		reporter := &recordingReporter{}
		h := handler{reporter: reporter}
		c, rr := test.NewContext(tt, nil)

		h.handleError(c, echo.ErrBadRequest.Wrap(errors.New("bad body")))

		assert.Equal(tt, http.StatusBadRequest, rr.Code)
		assert.Empty(tt, reporter.reported)
	})
}

func TestHandler(t *testing.T) {
	t.Run("surfaces generic errors as internal server errors", func(tt *testing.T) {
		h := handler{reporter: mockSentryClient{}}
		c, rr := test.NewContext(t, nil)
		err := errors.New("foo")

		h.handleError(c, err)

		assert.Equal(tt, http.StatusInternalServerError, rr.Code, "expected generic errors to be 500s")
		assert.Contains(tt, rr.Body.String(), "Internal Server Error", "expected generic errors to have the correct message")
	})

	t.Run("surfaces HTTP errors transparently", func(tt *testing.T) {
		h := handler{reporter: mockSentryClient{}}
		c, rr := test.NewContext(t, nil)
		err := echo.NewHTTPError(http.StatusForbidden, "")

		h.handleError(c, err)

		assert.Equal(tt, http.StatusForbidden, rr.Code, "expected HTTP errors to be correct")
		assert.Contains(tt, rr.Body.String(), "Forbidden", "expected HTTP errors to have the correct message")
	})

	t.Run("obfuscates custom HTTP error messages", func(tt *testing.T) {
		h := handler{reporter: mockSentryClient{}}
		c, rr := test.NewContext(t, nil)
		err := echo.NewHTTPError(http.StatusForbidden, "this should not be shown")

		h.handleError(c, err)

		assert.Equal(tt, http.StatusForbidden, rr.Code, "expected HTTP errors to be correct")
		assert.Contains(tt, rr.Body.String(), "Forbidden", "expected HTTP errors to have the correct message")
	})

	t.Run("overwrites default HTTP status codes message when customErrorMessages is enabled", func(tt *testing.T) {
		h := handler{reporter: mockSentryClient{}, customErrorMessages: true}
		c, rr := test.NewContext(t, nil)
		err := echo.NewHTTPError(http.StatusBadRequest, "custom error message")

		h.handleError(c, err)

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
