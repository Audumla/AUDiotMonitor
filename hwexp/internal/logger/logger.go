package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Level string

const (
	Debug Level = "debug"
	Info  Level = "info"
	Warn  Level = "warn"
	Error Level = "error"
	Fatal Level = "fatal"
)

type Entry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     Level                  `json:"level"`
	Component string                 `json:"component"`
	Event     string                 `json:"event"`
	Message   string                 `json:"message"`
	Host      string                 `json:"host"`
	RequestID string                 `json:"request_id,omitempty"`
	ErrorCode string                 `json:"error_code,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

type Logger struct {
	host      string
	component string
}

func New(host, component string) *Logger {
	return &Logger{
		host:      host,
		component: component,
	}
}

func (l *Logger) log(level Level, event, message string, errorCode string, details map[string]interface{}) {
	entry := Entry{
		Timestamp: time.Now().UTC(),
		Level:     level,
		Component: l.component,
		Event:     event,
		Message:   message,
		Host:      l.host,
		ErrorCode: errorCode,
		Details:   details,
	}

	data, _ := json.Marshal(entry)
	fmt.Fprintln(os.Stderr, string(data))

	if level == Fatal {
		os.Exit(1)
	}
}

func (l *Logger) Debug(event, message string, details map[string]interface{}) {
	l.log(Debug, event, message, "", details)
}

func (l *Logger) Info(event, message string, details map[string]interface{}) {
	l.log(Info, event, message, "", details)
}

func (l *Logger) Warn(event, message string, details map[string]interface{}) {
	l.log(Warn, event, message, "", details)
}

func (l *Logger) Error(event, message, errorCode string, details map[string]interface{}) {
	l.log(Error, event, message, errorCode, details)
}

func (l *Logger) Fatal(event, message, errorCode string, details map[string]interface{}) {
	l.log(Fatal, event, message, errorCode, details)
}
