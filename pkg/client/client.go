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

type sentry struct {
	client       errorReporter
	filterFields []string
}

// Sentry wraps Sentry's raven.Sentry.
type Sentry interface {
	Report(error, *http.Request)
}

// New returns an instance of Client
func New(env, dsn string) (Sentry, error) {
	defaultTags := map[string]string{
		"environment": env,
		"release":     os.Getenv("RELEASE"),
	}

	client, err := raven.NewWithTags(dsn, defaultTags)
	if err != nil {
		return &sentry{}, errors.Wrap(err, "client")
	}

	return &sentry{client: client}, nil
}

// Report sends information to Sentry.
func (c *sentry) Report(err error, req *http.Request) {
	stacktrace := raven.NewException(err, raven.GetOrNewStacktrace(err, 0, 2, nil))
	httpContext := raven.NewHttp(c.sanitizeRequest(req))
	packet := raven.NewPacket(err.Error(), stacktrace, httpContext)

	c.client.Capture(packet, nil)
}

func (c *sentry) sanitizeRequest(req *http.Request) *http.Request {
	url := req.URL
	query := url.Query()

	for _, keyword := range c.filterFields {
		for field := range url.Query() {
			if keyword == field {
				query[field] = []string{"[CENSORED]"}
			}
		}
	}

	req.URL.RawQuery = query.Encode()

	return req
}
