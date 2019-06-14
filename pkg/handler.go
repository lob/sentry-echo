package sentryecho

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo"
	logger "github.com/lob/logger-go"
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
func (h *handler) handleError(err error, c echo.Context) {
	log := getLogger(c)

	code := http.StatusInternalServerError
	msg := http.StatusText(code)

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		if h.customErrorMessages {
			msg = fmt.Sprintf("%v", he.Message)
		} else {
			msg = http.StatusText(code)
		}
	}

	if code == http.StatusInternalServerError {
		h.reporter.Report(err, c.Request())
	}

	log.Root(logger.Data{"status_code": code}).Err(err).Error("request error")

	err = c.JSON(code, map[string]interface{}{"error": map[string]interface{}{"message": msg, "status_code": code}})
	if err != nil {
		log.Err(err).Error("error handler json error")
	}
}
