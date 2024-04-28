// Package logger is used for logging. The default one pushes structured (JSON) logs.
// This is a barebones logger which I use and have not required any other logging libraries till date.
// It depends on your hosting environment and other complex requirements with logging.
package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"
)

func init() {
	_ = Logger(defaultLogger)
}

const (
	// LogTypeInfo is for logging type 'info'
	LogTypeInfo = "info"
	// LogTypeWarn is for logging type 'warn'
	LogTypeWarn = "warn"
	// LogTypeError is for logging type 'error'
	LogTypeError = "error"
	// LogTypeFatal is for logging type 'fatal'
	LogTypeFatal = "fatal"
)

// Logger interface defines all the logging methods to be implemented
type Logger interface {
	Info(ctx context.Context, payload ...any)
	Warn(ctx context.Context, payload ...any)
	Error(ctx context.Context, payload ...any)
	Fatal(ctx context.Context, payload ...any)
}

// LogHandler implements Logger
type LogHandler struct {
	Skipstack  int
	appName    string
	appVersion string
	params     map[string]string
}

func (lh *LogHandler) defaultPayload(severity string) map[string]any {
	_, file, line, _ := runtime.Caller(lh.Skipstack)
	payload := map[string]any{
		"app":        lh.appName,
		"appVersion": lh.appVersion,
		"severity":   severity,
		"at":         fmt.Sprintf("%s:%d", file, line),
		"timestamp":  time.Now(),
	}
	for key, value := range lh.params {
		payload[key] = value
	}
	return payload
}

func (lh *LogHandler) serialize(severity string, data ...any) (string, error) {
	payload := lh.defaultPayload(severity)
	for idx, value := range data {
		payload[fmt.Sprintf("%d", idx)] = fmt.Sprintf("%+v", value)
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (lh *LogHandler) log(severity string, payload ...any) {
	out, err := lh.serialize(severity, payload...)
	if err != nil {
		fmt.Printf("%+v\n", err)
		return
	}

	switch severity {
	case LogTypeFatal:
		{
			fmt.Println(out)
			os.Exit(1)
		}
	}
	fmt.Println(out)
}

// Info is for logging items with severity 'info'
func (lh *LogHandler) Info(ctx context.Context, payload ...any) {
	lh.log(LogTypeInfo, payload...)
}

// Warn is for logging items with severity 'Warn'
func (lh *LogHandler) Warn(ctx context.Context, payload ...any) {
	lh.log(LogTypeWarn, payload...)
}

// Error is for logging items with severity 'Error'
func (lh *LogHandler) Error(ctx context.Context, payload ...any) {
	lh.log(LogTypeError, payload...)
}

// Fatal is for logging items with severity 'Fatal'
func (lh *LogHandler) Fatal(ctx context.Context, payload ...any) {
	lh.log(LogTypeFatal, payload...)
}

// New returns a new instance of LogHandler
func New(
	appname string,
	appversion string,
	skipStack uint8,
	params map[string]string,
) *LogHandler {
	if skipStack <= 1 {
		skipStack = 4
	}

	return &LogHandler{
		Skipstack:  int(skipStack),
		appName:    appname,
		appVersion: appversion,
		params:     params,
	}
}
