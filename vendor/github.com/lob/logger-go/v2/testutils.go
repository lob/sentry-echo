package logger

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type TestWriter struct {
	wrote  int
	closed bool
	msg    string
}

func NewTestWriter() *TestWriter {
	return &TestWriter{
		wrote:  0,
		closed: false,
	}
}

func (fl *TestWriter) Write(b []byte) (int, error) {
	if fl.closed {
		return 0, errors.New("can't write to a closed writer")
	}

	fl.msg = string(b)
	fl.wrote++
	return 0, nil
}

func (fl *TestWriter) Close() error {
	fl.closed = true
	return nil
}

func shouldPanic(t *testing.T, f func()) {
	defer func() { recover() }()
	f()
	t.Errorf("should have panicked")
}

func testLogger(t *testing.T, infoLevel string, infoMsg string, global bool) {
	var id string
	var data Data
	var log Logger
	rootData := Data{"r1": "test", "r2": "moreTest"}

	origStdout := os.Stdout
	defer func() {
		os.Stdout = origStdout
		defaultLogger = New("")
	}()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("unexpected error when creating a pipe: %s", err)
	}
	os.Stdout = w

	if global {
		defaultLogger = New("")
	} else {
		id = "testId"
		data = Data{"data": "test"}
		var e error
		if infoLevel == "error" {
			e = errors.New("pkg error")
		} else {
			e = fmt.Errorf("runtime error")
		}
		log = New("").ID(id).Err(e).Data(data).Data(data).Root(rootData).Root(rootData)
	}

	d1, d2, d3, d4 :=
		Data{"1": "1"},
		Data{"2": 2},
		Data{"3": []int{3, 4, 5}},
		Data{"4": Data{"5": 6.5}}

	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	if global {
		switch infoLevel {
		case "error":
			Error(infoMsg, d1, d2, d3, d4)
		case "warn":
			Warn(infoMsg, d1, d2, d3, d4)
		case "debug":
			Debug(infoMsg, d1, d2, d3, d4)
		case "info":
			Info(infoMsg, d1, d2, d3, d4)
		case "trace":
			Trace(infoMsg, d1, d2, d3, d4)
		case "panic":
			shouldPanic(t, func() {
				Panic(infoMsg, d1, d2, d3, d4)
			})
		}
	} else {
		switch infoLevel {
		case "error":
			log.Error(infoMsg, d1, d2, d3, d4)
		case "warn":
			log.Warn(infoMsg, d1, d2, d3, d4)
		case "debug":
			log.Debug(infoMsg, d1, d2, d3, d4)
		case "info":
			log.Info(infoMsg, d1, d2, d3, d4)
		case "trace":
			log.Trace(infoMsg, d1, d2, d3, d4)
		case "panic":
			shouldPanic(t, func() {
				log.Panic(infoMsg, d1, d2, d3, d4)
			})
		}
	}

	if err = w.Close(); err != nil {
		t.Fatalf("unexpected error closing write pipe: %s", err)
	}

	logLine := <-outC

	if global {
		assert.Contains(t, logLine, fmt.Sprintf(`"level":"%s"`, infoLevel))
		assert.Contains(t, logLine, `"host":`)
		assert.Contains(t, logLine, `"release":`)
		assert.Contains(t, logLine, `"nanoseconds":`)
		assert.Contains(t, logLine, `"timestamp":`)
		assert.Contains(t, logLine, `"data":{"1":"1","2":2,"3":[3,4,5],"4":{"5":6.5}}`)
	} else {
		assert.Contains(t, logLine, fmt.Sprintf(`"id":"%s"`, id))
		assert.Contains(t, logLine, `"error":`)
		assert.Contains(t, logLine, fmt.Sprintf(`"level":"%s"`, infoLevel))
		assert.Contains(t, logLine, fmt.Sprintf(`"status":"%s"`, infoLevel))
		assert.Contains(t, logLine, `"host":`)
		assert.Contains(t, logLine, `"service":`)
		assert.Contains(t, logLine, `"release":`)
		assert.Contains(t, logLine, `"nanoseconds":`)
		assert.Contains(t, logLine, `"timestamp":`)
		assert.Contains(t, logLine, `"r1":"test"`)
		assert.Contains(t, logLine, `"r2":"moreTest"`)
		assert.Contains(t, logLine, `"data":{"1":"1","2":2,"3":[3,4,5],"4":{"5":6.5},"data":"test"}`)
	}
}
