package log

import (
	"strings"
	"testing"
)

type FakeWriter struct {
	WasUsed bool
}

// Returns and clears the status flag for the FakeWriter
func (w *FakeWriter) Check() bool {
	status := w.WasUsed
	w.WasUsed = false
	return status
}

// Implents io.Writer and simply sets a boolean as to whether or not we were 
// called so we can verify our filtering of lower log levels works. 
func (w *FakeWriter) Write(p []byte) (n int, err error) {
	w.WasUsed = true
	return
}

type RecordWriter struct {
	LastWrite []byte // The entire contents of the most recent Write()
}

func (w *RecordWriter) Write(p []byte) (n int, err error) {
	w.LastWrite = p
	return len(p), nil
}

// Cycles through all log levels and ensures that a lower level messages does
// not get logged.
func TestLevels(t *testing.T) {
	w := &FakeWriter{}
	var sysLevel, msgLevel LogLevel
	l, err := NewLevel(EMERG, true, w, "", Lshortfile|Ldate|Lmicroseconds)
	if err != nil {
		t.Error(err)
	}
	for sysLevel = 0; sysLevel.Int() < len(LevelNames); sysLevel++ {
		l.Level = sysLevel
		for msgLevel = 0; msgLevel.Int() < len(LevelNames); msgLevel++ {
			l.prefixOutput(msgLevel, "Log this!")
			if w.Check() && msgLevel.Int() > sysLevel.Int() {
				t.Error("Logging set to:", sysLevel,
					"but a message got through at level:", msgLevel)
			}
		}
	}
}

// TestString checks that LogLevel.String() is functioning properly.
func TestString(t *testing.T) {
	w := &RecordWriter{}
	l, err := NewLevel(EMERG, true, w, "", 0)
	if err != nil {
		t.Error(err)
	}
	for i := 0; i <= maxLevel; i++ {
		l.prefixOutput(LogLevel(i), "")
		if strings.Contains(string(w.LastWrite), "INVALID ") {
			t.Error("Got INVALID prefix for log level:", i)
		}
	}
}
