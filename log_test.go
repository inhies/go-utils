package log

import (
	"io/ioutil"
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

func TestChannels(t *testing.T) {
	w := ioutil.Discard
	l, err := NewLevel(EMERG, true, w, "", Lshortfile|Ldate|Lmicroseconds)
	if err != nil {
		t.Error(err)
	}

	// TODO(inhies): Check to make sure a race condition isn't possible with
	// these go routines. I think it might be but I haven't been able to prove
	// it.

	// Make sure we only receive messages of the current level
	lChan := make(chan Message)
	go func() {
		for {
			msg := <-lChan
			if msg.Level > l.Level {
				t.Error("Logging set to:", l.Level,
					"but a message got sent on channel at level:", msg.Level)
			}
		}
	}()

	// Make sure we receive all messages
	aChan := make(chan Message)
	var msgsRecvd int
	go func() {
		for {
			_ = <-aChan
			msgsRecvd++
		}
	}()

	// Register our channels to receive the log messages
	l.Split(aChan, true)  // Send all messages
	l.Split(lChan, false) // Send only messages >= l.Leve

	var sysLevel, msgLevel LogLevel
	var count int
	for sysLevel = 0; sysLevel.Int() < len(LevelNames); sysLevel++ {
		l.Level = sysLevel
		for msgLevel = 0; msgLevel.Int() < len(LevelNames); msgLevel++ {
			l.prefixOutput(msgLevel, "Log this!")
			count++
		}
	}

	// Make sure that all messages were sent to aChan
	if count != msgsRecvd {
		t.Error("We sent", count, "messages but received", msgsRecvd)
	}
}

// TestString checks that LogLevel.String() is functioning properly.
func TestString(t *testing.T) {
	w := &RecordWriter{}
	l, err := NewLevel(EMERG, true, w, "", 0)
	if err != nil {
		t.Error(err)
	}

	// Check all of the valid log levels.
	for i := 0; i <= maxLevel; i++ {
		l.prefixOutput(LogLevel(i), "")
		if strings.Contains(string(w.LastWrite), "INVALID") {
			t.Error("Got INVALID prefix for log level:", i)
		}
	}

	// Check some invalid log levels.
	l.prefixOutput(LogLevel(-1), "")
	if !strings.Contains(string(w.LastWrite), "INVALID") {
		t.Error("Invalid log level resulted in non-INVALID prefix")
	}
	l.prefixOutput(LogLevel(maxLevel+1), "")
	if !strings.Contains(string(w.LastWrite), "INVALID") {
		t.Error("Invalid log level resulted in non-INVALID prefix")
	}
}
