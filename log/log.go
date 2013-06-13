// Package log allows the use of BSD style log levels which allows you to be
// gracious with your use of log statements during development and then set the
// log level higher during production for less noise.
package log

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

// Package version
const (
	Version = "1.5"
)

// LogLevel represents the minimum level of messages we want to process.
type LogLevel int

// Implements the standard log.Logger as well as tracks what log level we
// currently want to use.
type Logger struct {
	// The lowest level of messages the logger will output
	Level LogLevel

	// Include the message level in logger output
	IncludeLevel bool

	// Timeout for writing log messages to a channel. Defaults to 1 second.
	Timeout time.Duration

	// Contains the number of messages that failed to send on channels due to
	// timeouts
	MissedMessages int

	// Slice containing channels that will only receive messages of Level and
	// higher
	levelChannels []chan Message

	// Slice containing channels that will receive all messages, regardless of
	// Level
	allChannels []chan Message

	// Standard Go log fields
	log.Logger
}

// Message represents a single log message that will be sent on a channel
// registered with the Split() method.
type Message struct {
	Level     LogLevel  // The level of the message
	Message   string    // The content of the message, represented as a string
	Timestamp time.Time // Timestamp of when the message was received
}

// String values for each of the log levels
var (
	LevelNames = [8]string{
		"EMERG",
		"ALERT",
		"CRIT",
		"ERR",
		"WARNING",
		"NOTICE",
		"INFO",
		"DEBUG",
	}
)

var (
	// maxLevel is the greatest valid LogLevel; anything strictly greater than
	// maxLevel is out of bounds.
	maxLevel = len(LevelNames) - 1
)

// Package values for each log level
const (
	NULL  = iota - 1 // NULL will be -1, so that it discards all output.
	EMERG            // EMERGE will be 0, and so on.
	ALERT
	CRIT
	ERR
	WARNING
	NOTICE
	INFO
	DEBUG

	// Imported from Golang log package:
	// Bits or'ed together to control what's printed. There is no control over the
	// order they appear (the order listed here) or the format they present (as
	// described in the comments).  A colon appears after these items:
	//	2009/01/23 01:23:23.123123 /a/b/c/d.go:23: message
	Ldate         = log.Ldate         // the date: 2009/01/23
	Ltime         = log.Ltime         // the time: 01:23:23
	Lmicroseconds = log.Lmicroseconds // microsecond resolution: 01:23:23.123123.  assumes Ltime.
	Llongfile     = log.Llongfile     // full file name and line number: /a/b/c/d.go:23
	Lshortfile    = log.Lshortfile    // final file name element and line number: d.go:23. overrides Llongfile
	LstdFlags     = log.LstdFlags     // initial values for the standard logger
)

// Common error messages.
var (
	InvalidLogLevelError = errors.New("log level invalid or out of bounds")
)

func (v LogLevel) String() string {
	// If v is out of bounds, return INVALID. Otherwise, look up the
	// string.
	if v.Int() < 0 || v.Int() > maxLevel {
		return "INVALID"
	}
	return LevelNames[v]
}

func (v LogLevel) Int() int {
	return int(v)
}

// Create a new logger with DEBUG as the default level. This is for backwards
// compatibility with exisiting code not utilizing go-log.
func New(out io.Writer, prefix string, flag int) (newLogger *Logger) {
	// Note that the false prevents the log from printing the urgency prefix, so
	// that it behaves exactly like stdlib logs.
	return &Logger{DEBUG, false, 1 * time.Second, 0, nil, nil, *log.New(out, prefix, flag)}
}

// Create a new logger with the specified level.
func NewLevel(level LogLevel, inc bool, out io.Writer, prefix string, flag int) (newLogger *Logger, err error) {

	if level.Int() > maxLevel {
		return nil, InvalidLogLevelError
	}

	newLogger = &Logger{level, inc, 1 * time.Second, 0, nil, nil, *log.New(out, prefix, flag)}
	return
}

// Accepts a string with the level name or an int corresponding to the level and
// returns the correct level.
func ParseLevel(input interface{}) (level LogLevel, err error) {
	n := NULL // Let NULL be the default.
	switch t := input.(type) {
	case string:
		// If the input is a string, then check if it matches any of the
		// LevelNames.
		tu := strings.ToUpper(t)
		for i, name := range LevelNames {
			if name == tu {
				return LogLevel(i), nil
			}
		}
		// If it doesn't match, then return NULL and an error.
		return NULL, InvalidLogLevelError
	case float64:
		// If t is out of bounds, then set the error and let the function return
		// as normal.
		if t > float64(maxLevel) || t < float64(0) {
			err = InvalidLogLevelError
		}
		n = int(t)
	case int:
		// As above, if t is out of bounds, set the error and return as normal.
		if t > maxLevel || t < 0 {
			err = InvalidLogLevelError
		}
		n = t
	}
	return LogLevel(n), err
}

// Split accepts a channel that will receive log messages in addition to them
// being sent to the logger's io.Writer. If sendAll is true then all messages,
// regardless of the configured logging level, will be sent to the channel.
func (logger *Logger) Split(c chan Message, sendAll bool) {
	if sendAll {
		logger.allChannels = append(logger.allChannels, c)
	} else {
		logger.levelChannels = append(logger.levelChannels, c)
	}
}

// prefixOutput obeys advanced logging rules and prepends prefixes before
// passing the final message to logger.Output().
func (logger *Logger) prefixOutput(level LogLevel, msg string) {
	// Send the message to channels that want all messages
	for _, c := range logger.allChannels {
		select {
		case c <- Message{level, msg, time.Now()}:
		case <-time.After(logger.Timeout):
			logger.MissedMessages++
		}
	}

	// Return if the message level isn't high enough
	if level > logger.Level {
		return
	}

	// Send the message to channels that only want messages of certain levels
	for _, c := range logger.levelChannels {
		select {
		case c <- Message{level, msg, time.Now()}:
		case <-time.After(logger.Timeout):
			logger.MissedMessages++
		}
	}

	if logger.IncludeLevel {
		// If we should include the level, prepend it.
		logger.Output(3, level.String()+" "+msg)
	} else { // Otherwise, give the message without any modifications.
		logger.Output(3, msg)
	}
}

func (logger *Logger) Debug(v ...interface{}) {
	logger.prefixOutput(DEBUG, fmt.Sprint(v...))
}

func (logger *Logger) Info(v ...interface{}) {
	logger.prefixOutput(INFO, fmt.Sprint(v...))
}

func (logger *Logger) Notice(v ...interface{}) {
	logger.prefixOutput(NOTICE, fmt.Sprint(v...))
}

func (logger *Logger) Warning(v ...interface{}) {
	logger.prefixOutput(WARNING, fmt.Sprint(v...))
}

func (logger *Logger) Err(v ...interface{}) {
	logger.prefixOutput(ERR, fmt.Sprint(v...))
}

func (logger *Logger) Crit(v ...interface{}) {
	logger.prefixOutput(CRIT, fmt.Sprint(v...))
}

func (logger *Logger) Alert(v ...interface{}) {
	logger.prefixOutput(ALERT, fmt.Sprint(v...))
}

func (logger *Logger) Emerg(v ...interface{}) {
	logger.prefixOutput(EMERG, fmt.Sprint(v...))
}

func (logger *Logger) Fatal(v ...interface{}) {
	logger.prefixOutput(EMERG, fmt.Sprint(v...))
	os.Exit(1)
}

func (logger *Logger) Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	logger.prefixOutput(EMERG, s)
	panic(s)
}

func (logger *Logger) Debugln(v ...interface{}) {
	logger.prefixOutput(DEBUG, fmt.Sprintln(v...))
}

func (logger *Logger) Infoln(v ...interface{}) {
	logger.prefixOutput(INFO, fmt.Sprintln(v...))
}

func (logger *Logger) Noticeln(v ...interface{}) {
	logger.prefixOutput(NOTICE, fmt.Sprintln(v...))
}

func (logger *Logger) Warningln(v ...interface{}) {
	logger.prefixOutput(WARNING, fmt.Sprintln(v...))
}

func (logger *Logger) Errln(v ...interface{}) {
	logger.prefixOutput(ERR, fmt.Sprintln(v...))
}

func (logger *Logger) Critln(v ...interface{}) {
	logger.prefixOutput(CRIT, fmt.Sprintln(v...))
}

func (logger *Logger) Alertln(v ...interface{}) {
	logger.prefixOutput(ALERT, fmt.Sprintln(v...))
}

func (logger *Logger) Emergln(v ...interface{}) {
	logger.prefixOutput(EMERG, fmt.Sprintln(v...))
}

func (logger *Logger) Fatalln(v ...interface{}) {
	logger.prefixOutput(EMERG, fmt.Sprintln(v...))
	os.Exit(1)
}

func (logger *Logger) Panicln(v ...interface{}) {
	s := fmt.Sprintln(v...)
	logger.prefixOutput(EMERG, s)
	panic(s)
}

func (logger *Logger) Debugf(format string, v ...interface{}) {
	logger.prefixOutput(DEBUG, fmt.Sprintf(format, v...))
}

func (logger *Logger) Infof(format string, v ...interface{}) {
	logger.prefixOutput(INFO, fmt.Sprintf(format, v...))
}

func (logger *Logger) Noticef(format string, v ...interface{}) {
	logger.prefixOutput(NOTICE, fmt.Sprintf(format, v...))
}

func (logger *Logger) Warningf(format string, v ...interface{}) {
	logger.prefixOutput(WARNING, fmt.Sprintf(format, v...))
}

func (logger *Logger) Errf(format string, v ...interface{}) {
	logger.prefixOutput(ERR, fmt.Sprintf(format, v...))
}

func (logger *Logger) Critf(format string, v ...interface{}) {
	logger.prefixOutput(CRIT, fmt.Sprintf(format, v...))
}

func (logger *Logger) Alertf(format string, v ...interface{}) {
	logger.prefixOutput(ALERT, fmt.Sprintf(format, v...))
}

func (logger *Logger) Emergf(format string, v ...interface{}) {
	logger.prefixOutput(EMERG, fmt.Sprintf(format, v...))
}

func (logger *Logger) Fatalf(format string, v ...interface{}) {
	logger.prefixOutput(EMERG, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func (logger *Logger) Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	logger.prefixOutput(EMERG, s)
	panic(s)
}
