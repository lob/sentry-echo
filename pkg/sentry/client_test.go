package sentry

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
	t.Run("create client", func(tt *testing.T) {
		sentry, err := New("")

		assert.NoError(tt, err)
		assert.NotNil(tt, sentry)
	})

	t.Run("default filter headers", func(tt *testing.T) {
		sentry, err := New("")
		assert.NoError(tt, err)

		assert.NotEmpty(tt, sentry.options.FilteredHeaders)
	})
}

func TestNewWithOptions(t *testing.T) {
	opts := Options{DSN: "", FilteredFields: []string{"sekrit"}}
	sentry, err := NewWithOptions(opts)

	assert.NoError(t, err)
	assert.NotZero(t, sentry)
	assert.Equal(t, opts, sentry.options)
}

func TestReport(t *testing.T) {
	t.Run("report without request", func(tt *testing.T) {
		m := &mockReporter{}
		c := &Sentry{client: m}

		errString := "some error"

		c.Report(errors.New(errString), nil)

		assert.Equal(t, 1, m.numCallsToCapture)
		assert.Equal(t, errString, m.packet.Message)
	})

	t.Run("report with request", func(tt *testing.T) {
		m := &mockReporter{}
		c := &Sentry{client: m}

		errString := "some error"

		req := httptest.NewRequest("GET", "/path", strings.NewReader(`data`))
		c.Report(errors.New(errString), req)

		assert.Equal(t, 1, m.numCallsToCapture)
		assert.Equal(t, errString, m.packet.Message)
	})
}

func TestReportSantizesRequest(t *testing.T) {
	m := &mockReporter{}
	c := &Sentry{
		client: m,
		options: Options{
			FilteredFields:  []string{"sekrit"},
			FilteredHeaders: []string{"Authorization"},
		},
	}

	t.Run("sanitizes query string parameters", func(tt *testing.T) {
		req := httptest.NewRequest("GET", "/path?sekrit=ssssshhhhhh", strings.NewReader(`data`))
		c.Report(errors.New("aieeeeee"), req)

		reported, ok := m.packet.Interfaces[1].(*raven.Http)
		assert.True(tt, ok)
		q, _ := url.ParseQuery(reported.Query)
		assert.NotEqual(tt, "ssssshhhhhh", q.Get("sekrit"))
	})

	t.Run("sanitizes headers", func(tt *testing.T) {
		req := httptest.NewRequest("GET", "/path?sekrit=ssssshhhhhh", nil)
		req.Header.Set("Authorization", "sekrit")

		c.Report(errors.New("aieeeeee"), req)

		reported, ok := m.packet.Interfaces[1].(*raven.Http)
		assert.True(t, ok)
		assert.Equal(tt, CensoredValueReplacement, reported.Headers["Authorization"])
	})

	t.Run("censors cookies", func(tt *testing.T) {
		req := httptest.NewRequest("GET", "/path?sekrit=ssssshhhhhh", nil)
		req.AddCookie(&http.Cookie{Name: "session", Value: "token"})

		c.Report(errors.New("aieeeeee"), req)

		reported, ok := m.packet.Interfaces[1].(*raven.Http)
		assert.True(t, ok)
		assert.Equal(tt, CensoredValueReplacement, reported.Cookies)
		assert.Equal(tt, CensoredValueReplacement, reported.Headers["Cookie"])
	})
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
