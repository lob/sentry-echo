package sentryecho

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo"
	logger "github.com/lob/logger-go"
)

// HTTPErrorReporter defines an interface for reporting errors associated with a Request
type HTTPErrorReporter interface {
	Report(error, *http.Request) (string, chan error)
}

type handler struct {
	reporter HTTPErrorReporter
}

// RegisterErrorHandler registers an error reporter as the HTTP Error Handler for Echo
func RegisterErrorHandler(e *echo.Echo, reporter HTTPErrorReporter) {
	h := handler{reporter}

	e.HTTPErrorHandler = h.handleError
}

// handleError is an Echo error handler that uses HTTP errors accordingly, and any
// generic error will be interpreted as an internal server error.
func (h *handler) handleError(err error, c echo.Context) {
	log := getLogger(c)

	code := http.StatusInternalServerError
	msg := http.StatusText(code)

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		msg = fmt.Sprintf("%v", he.Message)
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
