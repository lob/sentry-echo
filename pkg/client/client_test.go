package client

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	raven "github.com/getsentry/raven-go"
	"github.com/stretchr/testify/assert"
)

type mockReporter struct {
	numCallsToCapture int
	packet            *raven.Packet
}

func (m *mockReporter) Capture(packet *raven.Packet, captureTags map[string]string) (eventID string, ch chan error) {
	m.numCallsToCapture++
	m.packet = packet
	return "", nil
}

func TestNew(t *testing.T) {
	sentry, err := New("")

	assert.NoError(t, err)
	assert.NotNil(t, sentry)
}

func TestNewWithOptions(t *testing.T) {
	opts := Options{DSN: "", FilteredFields: []string{"sekrit"}}
	sentry, err := NewWithOptions(opts)

	assert.NoError(t, err)
	assert.NotZero(t, sentry)
	assert.Equal(t, opts, sentry.options)
}

func TestReport(t *testing.T) {
	m := &mockReporter{}
	c := &Sentry{client: m}

	errString := "some error"

	req := httptest.NewRequest("GET", "/path", strings.NewReader(`data`))
	c.Report(errors.New(errString), req)

	assert.Equal(t, 1, m.numCallsToCapture)
	assert.Equal(t, errString, m.packet.Message)
}

func TestReportSantizesRequest(t *testing.T) {
	m := &mockReporter{}
	c := &Sentry{
		client: m,
		options: Options{
			FilteredFields: []string{"sekrit"},
		},
	}

	req := httptest.NewRequest("GET", "/path?sekrit=ssssshhhhhh", strings.NewReader(`data`))
	c.Report(errors.New("aieeeeee"), req)

	reported, ok := m.packet.Interfaces[1].(*raven.Http)
	assert.True(t, ok)
	q, _ := url.ParseQuery(reported.Query)
	assert.NotEqual(t, "ssssshhhhhh", q.Get("sekrit"))
}

func TestSantizeRequest(t *testing.T) {
	h := &Sentry{
		client: &mockReporter{},
		options: Options{
			FilteredFields: []string{"signature"},
		},
	}

	t.Run("censors secret query fields in request", func(tt *testing.T) {
		req := &http.Request{
			URL: &url.URL{
				RawQuery: "test=foo&signature=12345",
			},
		}

		sanitizedReq := h.sanitizeRequest(req)

		assert.Contains(tt, sanitizedReq.URL.Query()["test"], "foo")
		assert.Contains(tt, sanitizedReq.URL.Query()["signature"], "[CENSORED]")
	})
}
