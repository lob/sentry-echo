package sentry

import (
	"net/http"
	"os"

	raven "github.com/getsentry/raven-go"
	"github.com/pkg/errors"
)

// CensoredValueReplacement is the replacement string for filtered values
var CensoredValueReplacement = "[CENSORED]"

// DefaultFilteredHeaders is the set of headers filtered when creating the client via New()
var DefaultFilteredHeaders = []string{
	"Authorization",
}

type errorReporter interface {
	Capture(packet *raven.Packet, captureTags map[string]string) (eventID string, ch chan error)
}

// Options defines the configuration for a Sentry client
type Options struct {
	DSN             string
	FilteredFields  []string
	FilteredHeaders []string
}

// Sentry wraps Sentry's raven.Sentry.
type Sentry struct {
	options Options
	client  errorReporter
}

// New returns an instance of Sentry
func New(dsn string) (Sentry, error) {
	return NewWithOptions(Options{DSN: dsn, FilteredHeaders: DefaultFilteredHeaders})
}

// NewWithOptions returns an instance of Sentry with the specified Options
func NewWithOptions(options Options) (Sentry, error) {
	defaultTags := map[string]string{
		"environment": os.Getenv("ENVIRONMENT"),
		"release":     os.Getenv("RELEASE"),
	}

	client, err := raven.NewWithTags(options.DSN, defaultTags)
	if err != nil {
		return Sentry{}, errors.Wrap(err, "client")
	}

	return Sentry{
		options: options,
		client:  client,
	}, nil
}

// Report asynchronously sends information to Sentry. If your program ends before the http call is made
// then the error is lost.
func (c *Sentry) Report(err error, req *http.Request) (string, chan error) {
	captured := []raven.Interface{
		raven.NewException(err, raven.GetOrNewStacktrace(err, 0, 2, nil)),
	}
	if req != nil {
		captured = append(captured, raven.NewHttp(c.sanitizeRequest(req)))
	}

	packet := raven.NewPacket(err.Error(), captured...)

	return c.client.Capture(packet, nil)
}

// ReportSync synchronously sends information to Sentry.
func (c *Sentry) ReportSync(err error, req *http.Request) (string, error) {
	errorID, channel := c.Report(err, req)
	err = <-channel
	return errorID, err
}

func (c *Sentry) sanitizeRequest(req *http.Request) *http.Request {
	url := req.URL
	query := url.Query()

	// sanitize query string values
	for _, keyword := range c.options.FilteredFields {
		for field := range url.Query() {
			if keyword == field {
				query[field] = []string{CensoredValueReplacement}
			}
		}
	}

	// sanitized filtered headers
	for _, keyword := range c.options.FilteredHeaders {
		if _, ok := req.Header[keyword]; ok {
			req.Header[keyword] = []string{CensoredValueReplacement}
		}
	}

	// remove cookies
	if _, ok := req.Header["Cookie"]; ok {
		req.Header["Cookie"] = []string{CensoredValueReplacement}
	}

	req.URL.RawQuery = query.Encode()

	return req
}
