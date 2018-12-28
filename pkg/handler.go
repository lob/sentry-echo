package sentry

import (
	"net/http"

	raven "github.com/getsentry/raven-go"
	"github.com/labstack/echo"
	logger "github.com/lob/logger-go"

	"github.com/lob/sentry-echo/pkg/client"
)

type handler struct {
	reporter     client.ErrorReporter
	filterFields []string
}

// RegisterErrorHandler takes in an Echo router and registers routes onto it.
func RegisterErrorHandler(e *echo.Echo, reporter client.ErrorReporter, filterFields []string) {
	h := handler{reporter, filterFields}

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
		stacktrace := raven.NewException(err, raven.GetOrNewStacktrace(err, 0, 2, nil))
		httpContext := raven.NewHttp(h.sanitizeRequest(c.Request()))
		packet := raven.NewPacket(msg, stacktrace, httpContext)

		h.reporter.Capture(packet, map[string]string{})
	}

	log.Root(logger.Data{"status_code": code}).Err(err).Error("request error")

	err = c.JSON(code, map[string]interface{}{"error": map[string]interface{}{"message": msg, "status_code": code}})
	if err != nil {
		log.Err(err).Error("error handler json error")
	}
}

func (h *handler) sanitizeRequest(req *http.Request) *http.Request {
	url := req.URL
	query := url.Query()

	for _, keyword := range h.filterFields {
		for field := range url.Query() {
			if keyword == field {
				query[field] = []string{"[CENSORED]"}
			}
		}
	}

	req.URL.RawQuery = query.Encode()

	return req
}
