package logger

import (
	"io"
	"maps"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

type LoggerContext struct {
	id   string
	err  error
	data Data
	root Data
}

// Logger holds the zerolog logger and metadata.
type Logger struct {
	zl            zerolog.Logger
	loggerContext LoggerContext
}

// Data is a type alias so that it's much more concise to add additional data to
// log lines.
type Data = map[string]interface{}

// Option allows us to add fields to the output of the log without wrapping it in
// the data object.
type Option map[string]interface{}

// New prepares and creates a new Logger instance using the default Stdout writer.
func New(serviceName string, options ...Option) Logger {
	return NewWithWriter(serviceName, os.Stdout, options...)
}

// NewWithWriter prepares and creates a new Logger instance with a specified writer.
func NewWithWriter(serviceName string, w io.Writer, options ...Option) Logger {
	zerolog.TimestampFieldName = "timestamp"
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	lc := LoggerContext{
		data: Data{},
		root: Data{},
	}

	host, _ := os.Hostname()
	if serviceName == "" {
		serviceName = os.Getenv("SERVICE_NAME")
	}

	zl := zerolog.New(w).With().Timestamp()
	zl = zl.
		Str("host", host).
		Str("release", os.Getenv("RELEASE")).
		Str("service", serviceName).
		Str("name", serviceName)

	for _, option := range options {
		for key, value := range option {
			zl = zl.Str(key, value.(string))
		}
	}

	return Logger{
		zl:            zl.Logger(),
		loggerContext: lc,
	}
}

// ID returns a new Logger with the ID set to id.
func (log Logger) ID(id string) Logger {
	log.loggerContext.id = id
	return log
}

// Err returns a new Logger with the error set to err.
func (log Logger) Err(err error) Logger {
	log.loggerContext.err = err
	return log
}

// Data returns a new logger with the new data appended to the old list of data.
func (log Logger) Data(data Data) Logger {
	newData := Data{}
	for k, v := range log.loggerContext.data {
		newData[k] = v
	}
	for k, v := range data {
		newData[k] = v
	}
	log.loggerContext.data = newData
	return log
}

func (log Logger) data(data ...Data) Logger {
	for _, d := range data {
		log = log.Data(d)
	}
	return log
}

func (log Logger) GetData() Data {
	// send back a cloned one as maps are actually headers, and
	// the underlying data is actually a pointer
	return maps.Clone(log.loggerContext.data)
}

// ClearData returns a new logger with the data cleared.
func (log Logger) ClearData() Logger {
	log.loggerContext.data = Data{}
	return log
}

// Root returns a new logger with the root info appended to the old list of root
// info. This root info will be displayed at the top level of the log.
func (log Logger) Root(root Data) Logger {
	newRoot := Data{}
	for k, v := range log.loggerContext.root {
		newRoot[k] = v
	}
	for k, v := range root {
		newRoot[k] = v
	}
	log.loggerContext.root = newRoot
	return log
}

// ClearRoot returns a new logger with the root info cleared.
func (log Logger) ClearRoot() Logger {
	log.loggerContext.root = Data{}
	return log
}

// Info outputs an info-level log with a message and any additional data provided.
func (log Logger) Info(message string, fields ...Data) {
	sendLog(log, message, zerolog.InfoLevel, fields...)
}

// Error outputs an error-level log with a message and any additional data provided.
func (log Logger) Error(message string, fields ...Data) {
	sendLog(log, message, zerolog.ErrorLevel, fields...)
}

// Warn outputs a warn-level log with a message and any additional data provided.
func (log Logger) Warn(message string, fields ...Data) {
	sendLog(log, message, zerolog.WarnLevel, fields...)
}

// Debug outputs a debug-level log with a message and any additional data provided.
func (log Logger) Debug(message string, fields ...Data) {
	sendLog(log, message, zerolog.DebugLevel, fields...)
}

// Fatal outputs a fatal-level log with a message and any additional data provided.
// This will also call os.Exit(1)
func (log Logger) Fatal(message string, fields ...Data) {
	sendLog(log, message, zerolog.FatalLevel, fields...)
}

// Panic outputs a panic-level log with a message and any additional data provided.
func (log Logger) Panic(message string, fields ...Data) {
	sendLog(log, message, zerolog.PanicLevel, fields...)
}

// Trace outputs a trace-level log with a message and any additional data provided.
func (log Logger) Trace(message string, fields ...Data) {
	sendLog(log, message, zerolog.TraceLevel, fields...)
}

// GetLoggerContext returns the logger context for the logger.
func (log Logger) GetLoggerContext() *LoggerContext {
	return &log.loggerContext
}

func AddFields(loggerContext *LoggerContext, evt *zerolog.Event, level zerolog.Level) {
	evt.Int64("nanoseconds", zerolog.TimestampFunc().UnixNano())

	if level != zerolog.NoLevel {
		evt.Str("status", level.String())
	}

	if loggerContext.id != "" {
		evt.Str("id", loggerContext.id)
	}

	if len(loggerContext.root) > 0 {
		evt.Fields(loggerContext.root)
	}

	if len(loggerContext.data) > 0 {
		evt.Dict("data", zerolog.Dict().Fields(loggerContext.data))
	}

	if loggerContext.err != nil {
		evt.Stack().Err(loggerContext.err)
	}
}

func sendLog(log Logger, message string, level zerolog.Level, fields ...Data) {
	log = log.data(fields...)
	var event *zerolog.Event
	// panic/fatal levels using WithLevel don't behave like their actual
	// functions, so have to use them directly.
	switch level {
	case zerolog.PanicLevel:
		event = log.zl.Panic()
	case zerolog.FatalLevel:
		event = log.zl.Fatal()
	default:
		event = log.zl.WithLevel(level)
	}
	AddFields(&log.loggerContext, event, level)
	event.Msg(message)
}
