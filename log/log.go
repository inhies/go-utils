// Package log allows the use of BSD style log levels with the standard Go 'log' package
// This allows you to be gracious with your use of log statements during development
// and then set the log level higher during production for less noise.
package log

// TODO(inhies): Implement table based test for this. 
// TODO(inhies): Implement the log.Fatal functions

import (
	"fmt"
	"io"
	"log"
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
)

const (
	EMERG = iota
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

func lookup(input int) (word string) {
	if input > len(LevelNames)-1 || input < 0 {
		return "INVALID"
	}

	return LevelNames[input]
}

func (v LogLevel) String() string {
	return lookup(int(v))
}

func (v LogLevel) int() int {
	return int(v)
}

// Create a new logger with DEBUG as the default level. This is for backwards compatibility
// with exisiting code not utilizing go-utils/log
func New(out io.Writer, prefix string, flag int) (newLogger *Logger) {
	newLogger = &Logger{DEBUG, false, *log.New(out, prefix, flag)}
	return
}

// Create a new logger with the specified level
func NewLevel(level LogLevel, inc bool, out io.Writer, prefix string, flag int) (newLogger *Logger, err error) {
	if int(level) > len(LevelNames) -1 {
		err = fmt.Errorf("Invalid log level specified")
		return
	}

	newLogger = &Logger{level, inc, *log.New(out, prefix, flag)}
	return
}

// Accepts a string with the level name or an int corresponding to the level and returns the
// correct level.
func ParseLevel(input interface{}) (level LogLevel, err error) {
	switch t := input.(type) {
	case string:
		for i := 0; i < len(LevelNames); i++ {
			if LevelNames[i] == strings.ToUpper(t) {
				level = LogLevel(i)
				return
			}
		}
		err = fmt.Errorf("Unknown log level specified")
		return

	case float64:
		if t > float64(len(LevelNames)-1) || t < float64(0) {
			err = fmt.Errorf("Unknown log level specified")
			return
		}
		level = LogLevel(t)
	case int:
		if t > DEBUG {
			err = fmt.Errorf("Unknown log level specified")
			return
		}
		level = LogLevel(t)
	default:
		err = fmt.Errorf("Unknown log level specified")
		return
	}

	return
}

// Custom output function to enforce log levels before forwarding the message 
// to the log package
func (logger *Logger) MyOutput(level LogLevel, msg string) {
	if level > logger.Level {
		return
	}

	if logger.IncludeLevel {
		msg = logger.Level.String() + " " + msg
	}
	logger.Output(3, msg)
}

// Print style

func (logger *Logger) Debug(v ...interface{}) {
	logger.MyOutput(DEBUG, fmt.Sprint(v...))
}

func (logger *Logger) Info(v ...interface{}) {
	logger.MyOutput(INFO, fmt.Sprint(v...))
}

func (logger *Logger) Notice(v ...interface{}) {
	logger.MyOutput(NOTICE, fmt.Sprint(v...))
}

func (logger *Logger) Warning(v ...interface{}) {
	logger.MyOutput(WARNING, fmt.Sprint(v...))
}

func (logger *Logger) Err(v ...interface{}) {
	logger.MyOutput(ERR, fmt.Sprint(v...))
}

func (logger *Logger) Crit(v ...interface{}) {
	logger.MyOutput(CRIT, fmt.Sprint(v...))
}

func (logger *Logger) Alert(v ...interface{}) {
	logger.MyOutput(ALERT, fmt.Sprint(v...))
}

func (logger *Logger) Emerg(v ...interface{}) {
	logger.MyOutput(EMERG, fmt.Sprint(v...))
}

// Println style

func (logger *Logger) Debugln(v ...interface{}) {
	logger.MyOutput(DEBUG, fmt.Sprintln(v...))
}

func (logger *Logger) Infoln(v ...interface{}) {
	logger.MyOutput(INFO, fmt.Sprintln(v...))
}

func (logger *Logger) Noticeln(v ...interface{}) {
	logger.MyOutput(NOTICE, fmt.Sprintln(v...))
}

func (logger *Logger) Warningln(v ...interface{}) {
	logger.MyOutput(WARNING, fmt.Sprintln(v...))
}

func (logger *Logger) Errln(v ...interface{}) {
	logger.MyOutput(ERR, fmt.Sprintln(v...))
}

func (logger *Logger) Critln(v ...interface{}) {
	logger.MyOutput(CRIT, fmt.Sprintln(v...))
}

func (logger *Logger) Alertln(v ...interface{}) {
	logger.MyOutput(ALERT, fmt.Sprintln(v...))
}

func (logger *Logger) Emergln(v ...interface{}) {
	logger.MyOutput(EMERG, fmt.Sprintln(v...))
}

// Printf style

func (logger *Logger) Debugf(format string, v ...interface{}) {
	logger.MyOutput(DEBUG, fmt.Sprintf(format, v...))
}

func (logger *Logger) Infof(format string, v ...interface{}) {
	logger.MyOutput(INFO, fmt.Sprintf(format, v...))
}

func (logger *Logger) Noticef(format string, v ...interface{}) {
	logger.MyOutput(NOTICE, fmt.Sprintf(format, v...))
}

func (logger *Logger) Warningf(format string, v ...interface{}) {
	logger.MyOutput(WARNING, fmt.Sprintf(format, v...))
}

func (logger *Logger) Errf(format string, v ...interface{}) {
	logger.MyOutput(ERR, fmt.Sprintf(format, v...))
}

func (logger *Logger) Critf(format string, v ...interface{}) {
	logger.MyOutput(CRIT, fmt.Sprintf(format, v...))
}

func (logger *Logger) Alertf(format string, v ...interface{}) {
	logger.MyOutput(ALERT, fmt.Sprintf(format, v...))
}

func (logger *Logger) Emergf(format string, v ...interface{}) {
	logger.MyOutput(EMERG, fmt.Sprintf(format, v...))
}
