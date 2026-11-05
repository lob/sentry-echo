package sentryecho

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"
	logger "github.com/lob/logger-go/v2"
)

// Options defines a struct that allows you to modify the behavior of how errors are reported.
type Options struct {
	Reporter                  HTTPErrorReporter
	EnableCustomErrorMessages bool
}

// HTTPErrorReporter defines an interface for reporting errors associated with a Request.
type HTTPErrorReporter interface {
	Report(error, *http.Request) (string, chan error)
}

type handler struct {
	reporter            HTTPErrorReporter
	customErrorMessages bool
}

// RegisterErrorHandlerWithOptions registers an error reporter as the HTTP Error Handler
// for Echo and provides options on how the error should be treated.
func RegisterErrorHandlerWithOptions(e *echo.Echo, options Options) {
	h := handler{
		reporter:            options.Reporter,
		customErrorMessages: options.EnableCustomErrorMessages,
	}

	e.HTTPErrorHandler = h.handleError
}

// RegisterErrorHandler registers an error reporter as the HTTP Error Handler for Echo.
func RegisterErrorHandler(e *echo.Echo, reporter HTTPErrorReporter) {
	RegisterErrorHandlerWithOptions(e, Options{Reporter: reporter})
}

// handleError is an Echo error handler that uses HTTP errors accordingly, and any
// generic error will be interpreted as an internal server error.
func (h *handler) handleError(c *echo.Context, err error) {
	log := getLogger(c)

	// echo v5's own errors (ErrNotFound etc.) aren't *echo.HTTPError, so read the status through HTTPStatusCoder.
	code := echo.StatusCode(err)
	if code == 0 {
		code = http.StatusInternalServerError
	}
	msg := http.StatusText(code)

	var he *echo.HTTPError
	if h.customErrorMessages && errors.As(err, &he) && he.Message != "" {
		msg = he.Message
	}

	if code == http.StatusInternalServerError {
		h.reporter.Report(err, c.Request())
	}

	url := c.Request().URL
	log = log.Root(logger.Data{
		"status_code": code,
		"method":      c.Request().Method,
		"path":        url.Scheme + "://" + url.Host + url.Path,
	})
	if code >= 500 {
		// we only call .Err() if we want to include the stack trace in the log
		log = log.Err(err)
		log.Error("server error")
	} else {
		log.Warn("request error")
	}

	// format the error as JSON for the middleware response
	err = c.JSON(code, map[string]interface{}{"error": map[string]interface{}{"message": msg, "status_code": code}})
	if err != nil {
		log.Err(err).Error("error handler json error")
	}
}
