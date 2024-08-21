package tools

import (
	"fmt"
	"log"
	"os"
	"strings"
)

const (
	DEBUG = iota
	INFO
	WARN
	ERROR
	SILENT
)

var (
	LoggerOut   = os.Stdout
	Level       = INFO
	debugLogger *log.Logger
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
)

type Logger struct {
	Service string
	Name    string
}

func SetLogOutput(path string) {
	if path == "" {
		LoggerOut = os.Stdout
	} else {
		file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(fmt.Errorf("failed set log output file : %v", err))
		}
		LoggerOut = file
	}
}

func SetLevelStr(s string) {
	if s == "" {
		return
	}
	switch {
	case strings.EqualFold(s, "debug"):
		Level = DEBUG
	case strings.EqualFold(s, "info"):
		Level = INFO
	case strings.EqualFold(s, "warn"):
		Level = WARN
	case strings.EqualFold(s, "error"):
		Level = ERROR
	default:
		panic(fmt.Errorf("unknown log level: '%v'", s))
	}
}

func (l *Logger) Debug(f string, args ...any) {
	if debugLogger == nil {
		debugLogger = log.New(LoggerOut, "[DEBUG] ", log.Ldate|log.Ltime)
	}
	if Level <= DEBUG {
		debugLogger.Printf("[%s %s] %s", l.Service, l.Name, fmt.Sprintf(f, args...))
	}
}

func (l *Logger) Info(f string, args ...any) {
	if infoLogger == nil {
		infoLogger = log.New(LoggerOut, "[INFO] ", log.Ldate|log.Ltime)
	}
	if Level <= INFO {
		infoLogger.Printf("[%s %s] %s", l.Service, l.Name, fmt.Sprintf(f, args...))
	}
}

func (l *Logger) Warn(f string, args ...any) {
	if warnLogger == nil {
		warnLogger = log.New(LoggerOut, "[WARN] ", log.Ldate|log.Ltime)
	}
	if Level <= WARN {
		warnLogger.Printf("[%s %s] %s", l.Service, l.Name, fmt.Sprintf(f, args...))
	}
}

func (l *Logger) Error(f string, args ...any) {
	if errorLogger == nil {
		errorLogger = log.New(LoggerOut, "[ERROR] ", log.Ldate|log.Ltime)
	}
	if Level <= ERROR {
		errorLogger.Printf("[%s %s] %s", l.Service, l.Name, fmt.Sprintf(f, args...))
	}
}
