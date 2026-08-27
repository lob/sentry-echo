package logger

import (
	"io"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

// MiddlewareConfig can be used to configure the Echo Middleware.
type MiddlewareConfig struct {
	IsIgnorableError func(error) bool
	Writer           io.Writer
}

var defaultMiddlewareConfig = MiddlewareConfig{
	IsIgnorableError: func(err error) bool {
		return false
	},
}

const echoKey = "logger"

// Middleware attaches a Logger instance with a request ID onto the context. It
// also logs every request along with metadata about the request. To customize
// the middleware, use MiddlewareWithConfig.
func Middleware(serviceName string) func(next echo.HandlerFunc) echo.HandlerFunc {
	return MiddlewareWithConfig(serviceName, defaultMiddlewareConfig)
}

// MiddlewareWithConfig attaches a Logger instance with a request ID onto the
// context. It also logs every request along with metadata about the request.
// Pass in a MiddlewareConfig to customize the behavior of the middleware.
func MiddlewareWithConfig(serviceName string, opts MiddlewareConfig) func(next echo.HandlerFunc) echo.HandlerFunc {
	var l Logger
	if opts.Writer != nil {
		l = NewWithWriter(serviceName, opts.Writer)
	} else {
		l = New(serviceName)
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			t1 := time.Now()

			// create a request ID that will be attached to the
			// logger
			id, err := uuid.NewV4()
			if err != nil {
				return errors.WithStack(err)
			}

			log := l.ID(id.String()).Root(Data{
				"method":     c.Request().Method,
				"route":      c.Path(),
				"path":       c.Request().URL.Path,
				"trace_id":   c.Request().Header.Get("x-amzn-trace-id"),
				"referer":    c.Request().Referer(),
				"user_agent": c.Request().UserAgent(),
			})
			c.Set(echoKey, log)

			if err := next(c); err != nil {
				if opts.IsIgnorableError(err) {
					log.Err(err).Warn("ignored error")
					return err
				}

				c.Error(err)
			}

			t2 := time.Now()

			log.Root(Data{
				"status_code":   c.Response().Status,
				"response_time": t2.Sub(t1).Seconds() * 1000,
			}).Info("handled request")

			return nil
		}
	}
}

// FromEchoContext returns a Logger from the given echo.Context with the serviceNameOrLogger Data merged in.
// If there is no attached logger, then it will return either the logger passed in,
// or a new Logger instance with the service's name populated in the correct fields.
func FromEchoContext[T string | Logger](c echo.Context, serviceNameOrLogger T) Logger {
	suppliedLogger, hasSuppliedLogger := any(serviceNameOrLogger).(Logger)
	// check if there's a logger attached to the echo context first
	if log, ok := c.Get(echoKey).(Logger); ok {
		if hasSuppliedLogger {
			// and if there's a supplied logger, merge the data from the supplied logger into the echo context logger
			return log.Data(suppliedLogger.GetData())
		}
		// otherwise, if there's no supplied logger, just return the echo context logger
		return log
	}

	// if there's no echo context logger, then just return the supplied logger
	if hasSuppliedLogger {
		return suppliedLogger
	}

	// and finally, just create a new logger with the service name
	// if we were unable to use an existing logger
	return New(any(serviceNameOrLogger).(string))
}
