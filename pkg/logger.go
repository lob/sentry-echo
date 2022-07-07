package sentryecho

import (
	"github.com/labstack/echo/v4"
	logger "github.com/lob/logger-go"
)

const loggerContextKey = "logger"

func getLogger(c echo.Context) logger.Logger {
	if log, ok := c.Get(loggerContextKey).(logger.Logger); ok {
		return log
	}
	return logger.New("")
}
