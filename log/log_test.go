package log

import (
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
			l.MyOutput(msgLevel, "Log this!")
			if w.Check() && msgLevel.Int() > sysLevel.Int() {
				t.Error("Logging set to:", sysLevel,
					"but a message got through at level:", msgLevel)
			}
		}
	}
}
