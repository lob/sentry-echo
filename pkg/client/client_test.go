package client

import (
	"errors"
	"net/http/httptest"
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

func TestReport(t *testing.T) {
	m := &mockReporter{}
	c := &Sentry{m}

	errString := "some error"

	req := httptest.NewRequest("GET", "/path", strings.NewReader(`data`))
	c.Report(errors.New(errString), req)

	assert.Equal(t, 1, m.numCallsToCapture)
	assert.Equal(t, errString, m.packet.Message)
}

func TestNew(t *testing.T) {
	sentry, err := New("test", "")

	assert.NoError(t, err)
	assert.NotNil(t, sentry)
}
