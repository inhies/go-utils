// Package log allows the use of BSD style log levels with the standard Go 'log' package
// This allows you to be gracious with your use of log statements during development
// and then set the log level higher during production for less noise.
package log

// TODO(inhies): Implement table based test for this. 
// TODO(inhies): Implement the log.Fatal functions

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const (
	Version = "1.4"
)

type LogLevel int

// Implements the standard log.Logger as well as tracks what log 
// level we currently want to use
type Logger struct {
	Level        LogLevel
	IncludeLevel bool
	log.Logger
}

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

	// maxLevel is the greatest valid LogLevel; anything strictly greater than
	// maxLevel is out of bounds.
	maxLevel = len(LevelNames) - 1
)

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

var (
	InvalidLogLevelError = errors.New("log level invalid or out of bounds")
)

func (v LogLevel) String() string {
	// If v is in bounds, look up the string. Otherwise, return "INVALID".
	if v.Int() < 0 || v.Int() > maxLevel {
		return LevelNames[v]
	}
	return "INVALID"
}

func (v LogLevel) Int() int {
	return int(v)
}

// Create a new logger with DEBUG as the default level. This is for backwards
// compatibility with exisiting code not utilizing go-utils/log
func New(out io.Writer, prefix string, flag int) (newLogger *Logger) {
	// Note that the false prevents the log from printing the urgency prefix, so
	// that it behaves exactly like stdlib logs.
	return &Logger{DEBUG, false, *log.New(out, prefix, flag)}
}

// Create a new logger with the specified level
func NewLevel(level LogLevel, inc bool, out io.Writer, prefix string, flag int) (newLogger *Logger, err error) {

	if level.Int() > maxLevel {
		return nil, InvalidLogLevelError
	}

	newLogger = &Logger{level, inc, *log.New(out, prefix, flag)}
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

// prefixOutput obeys advanced logging rules and prepends prefixes before
// passing the final message to logger.Output().
func (logger *Logger) prefixOutput(level LogLevel, msg string) {
	if level > logger.Level {
		return
	}

	if logger.IncludeLevel {
		// If we should include the level, prepend it.
		logger.Output(3, level.String()+" "+msg)
	} else { // Otherwise, give the message without any modifications.
		logger.Output(3, msg)
	}
}

// Print() style

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

// Println() style

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

// Printf() style

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
