package client

import (
	"net/http"
	"os"

	raven "github.com/getsentry/raven-go"
	"github.com/pkg/errors"
)

type ErrorReporter interface {
	Capture(packet *raven.Packet, captureTags map[string]string) (eventID string, ch chan error)
}

// Sentry wraps Sentry's raven.Sentry.
type Sentry struct {
	client ErrorReporter
}

// Report sends information to Sentry.
func (c *Sentry) Report(err error, req *http.Request) {
	stacktrace := raven.NewException(err, raven.GetOrNewStacktrace(err, 0, 2, nil))
	httpContext := raven.NewHttp(req)
	packet := raven.NewPacket(err.Error(), stacktrace, httpContext)

	c.client.Capture(packet, nil)
}

// New returns an instance of Client
func New(env, dsn string) (Sentry, error) {
	defaultTags := map[string]string{
		"environment": env,
		"release":     os.Getenv("RELEASE"),
	}

	client, err := raven.NewWithTags(dsn, defaultTags)
	if err != nil {
		return Sentry{}, errors.Wrap(err, "client")
	}

	return Sentry{client}, nil
}
