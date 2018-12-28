package sentry

import (
	"net/http"

	"github.com/labstack/echo"
	logger "github.com/lob/logger-go"

	"github.com/lob/sentry-echo/pkg/client"
)

type handler struct {
	reporter client.Sentry
}

// RegisterErrorHandler takes in an Echo router and registers routes onto it.
func RegisterErrorHandler(e *echo.Echo, reporter client.Sentry) {
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
		msg = http.StatusText(code)
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
