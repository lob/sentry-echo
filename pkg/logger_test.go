package sentry

import (
	"testing"

	logger "github.com/lob/logger-go"
	"github.com/lob/sentry-echo/internal/test"
	"github.com/stretchr/testify/assert"
)

func TestGetLogger(t *testing.T) {

	t.Run("return new logger if not in context", func(tt *testing.T) {
		c, _ := test.NewContext(tt, nil)

		log := getLogger(c)
		assert.NotZero(tt, log)
	})

	t.Run("returns logger from Context", func(tt *testing.T) {
		log := logger.New()
		c, _ := test.NewContext(t, nil)
		c.Set(loggerContextKey, log)

		l := getLogger(c)

		assert.Equal(t, log, l)
	})
}
