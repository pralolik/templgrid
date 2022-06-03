package logging

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
)

var (
	// ErrParseStrToLevel indicates that string given to function ParseLevel can't be parsed to Level.
	ErrParseStrToLevel = errors.New("string can't be parsed to level, use: `error`, `info`, `warning`, `debug`")
)

// Level represents enumeration of logging levels.
type Level int

const (
	// ERR represents error logging level.
	ERR Level = iota
	// INF represents info logging level.
	INF
	// DBG represents debug logging level.
	DBG
	// DSB empty log.
	DSB
)

func (l Level) String() string { return [4]string{"Error", "Info", "Debug", "Disabled"}[l] }

// Logger represents logging logic.
type Logger interface {
	// Debug prints message with debug level.
	Debug(format string, v ...interface{})
	// Info prints message with info level.
	Info(format string, v ...interface{})
	// Error prints message with error level.
	Error(format string, v ...interface{})
}

func NewDisabledLog() Logger { return &DisabledLog{} }

type DisabledLog struct{}

func (l *DisabledLog) Debug(string, ...interface{}) {}
func (l *DisabledLog) Info(string, ...interface{})  {}
func (l *DisabledLog) Error(string, ...interface{}) {}

// NewStdLog returns a new instance of StdLog struct.
// Takes variadic options which will be applied to StdLog.
func NewStdLog(level Level) Logger {
	if level == DSB {
		return NewDisabledLog()
	}
	l := &StdLog{
		err: log.New(os.Stderr, "\033[31mERR\033[0m: ", log.Ldate|log.Ltime),
		inf: log.New(os.Stdout, "\033[32mINF\033[0m: ", log.Ldate|log.Ltime),
		dbg: log.New(os.Stdout, "\033[35mDBG\033[0m: ", log.Ldate|log.Ltime),
		lvl: level,
	}

	return l
}

// StdLog represents standard library logger with levels.
type StdLog struct {
	err, inf, dbg *log.Logger
	lvl           Level
}

func (l *StdLog) Debug(format string, v ...interface{}) {
	if l.lvl < DBG {
		return
	}
	l.dbg.Printf(format, v...)
}

func (l *StdLog) Info(format string, v ...interface{}) {
	if l.lvl < INF {
		return
	}
	l.inf.Printf(format, v...)
}

func (l *StdLog) Error(format string, v ...interface{}) {
	if l.lvl < ERR {
		return
	}
	l.err.Printf(format, v...)
}

func ParseLevel(lvl string) (Level, error) {
	levels := map[string]Level{
		strings.ToLower(ERR.String()): ERR,
		strings.ToLower(INF.String()): INF,
		strings.ToLower(DBG.String()): DBG,
		strings.ToLower(DSB.String()): DSB,
	}
	level, ok := levels[strings.ToLower(lvl)]
	if !ok {
		return INF, fmt.Errorf("%s %w", lvl, ErrParseStrToLevel)
	}
	return level, nil
}
