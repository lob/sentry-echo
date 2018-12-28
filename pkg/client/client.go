package client

import (
	"net/http"
	"os"

	raven "github.com/getsentry/raven-go"
	"github.com/pkg/errors"
)

type errorReporter interface {
	Capture(packet *raven.Packet, captureTags map[string]string) (eventID string, ch chan error)
}

// Options defines the configuration for a Sentry client
type Options struct {
	DSN            string
	FilteredFields []string
}

// Sentry wraps Sentry's raven.Sentry.
type Sentry struct {
	options Options
	client  errorReporter
}

// New returns an instance of Sentry
func New(dsn string) (Sentry, error) {
	return NewWithOptions(Options{DSN: dsn})
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

// Report sends information to Sentry.
func (c *Sentry) Report(err error, req *http.Request) {
	stacktrace := raven.NewException(err, raven.GetOrNewStacktrace(err, 0, 2, nil))
	httpContext := raven.NewHttp(c.sanitizeRequest(req))
	packet := raven.NewPacket(err.Error(), stacktrace, httpContext)

	c.client.Capture(packet, nil)
}

func (c *Sentry) sanitizeRequest(req *http.Request) *http.Request {
	url := req.URL
	query := url.Query()

	for _, keyword := range c.options.FilteredFields {
		for field := range url.Query() {
			if keyword == field {
				query[field] = []string{"[CENSORED]"}
			}
		}
	}

	req.URL.RawQuery = query.Encode()

	return req
}
