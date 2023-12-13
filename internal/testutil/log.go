package testutil

import (
	"bytes"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/sorintlab/errors"
)

// testingWriter is a WriteSyncer that writes to the given testing.TB.
type testingWriter struct {
	t *testing.T
}

func NewTestingWriter(t *testing.T) *testingWriter {
	return &testingWriter{t: t}
}

func (w *testingWriter) Write(p []byte) (n int, err error) {
	n = len(p)

	// Strip trailing newline because t.Log always adds one.
	p = bytes.TrimRight(p, "\n")

	// Note: t.Log is safe for concurrent use.
	w.t.Logf("%s", p)

	return n, nil
}

func NewLogger(t *testing.T) zerolog.Logger {
	detailedErrors, _ := strconv.ParseBool(os.Getenv("DETAILED_ERRORS"))

	if detailedErrors {
		zerolog.ErrorMarshalFunc = errors.ErrorMarshalFunc
	}

	cw := zerolog.ConsoleWriter{
		Out:                 NewTestingWriter(t),
		TimeFormat:          time.RFC3339Nano,
		FormatErrFieldValue: errors.FormatErrFieldValue,
	}

	zerolog.TimeFieldFormat = time.RFC3339Nano

	return zerolog.New(cw).With().Timestamp().Stack().Caller().Logger().Level(zerolog.InfoLevel).Output(cw)
}
