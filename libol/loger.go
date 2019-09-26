package libol

import (
    "log"
    "fmt"
)

type Logger struct {
    Level int
    Errors []string
}

func (this *Logger) Info(format string, v ...interface{}) {
    log.Printf(fmt.Sprintf("INFO %s", format), v...)
}

func (this *Logger) Warn(format string, v ...interface{}) {
    log.Printf(fmt.Sprintf("WARN %s", format), v...)
}

func (this *Logger) Error(format string, v ...interface{}) {
    //TODO insert to Errors.
    log.Printf(fmt.Sprintf("ERROR %s", format), v...)
}

func (this *Logger) Debug(format string, v ...interface{}) {
    log.Printf(fmt.Sprintf("DEBUG %s", format), v...)
}

func (this *Logger) Fatal(format string, v ...interface{}) {
    //TODO saved to Errors and Publish it.
    log.Printf(fmt.Sprintf("FATAL %s", format), v...)
}

func (this *Logger) Printf(format string, v ...interface{}) {
    log.Printf(fmt.Sprintf("PRINT %s", format), v...)
}

var Log = Logger {
    Level: 1,
    Errors: make([]string, 1024),
}

func Error(format string, v ...interface{}) {
    Log.Error(format, v...)
}

func Debug(format string, v ...interface{}) {
    Log.Debug(format, v...)
}

func Info(format string, v ...interface{}) {
    Log.Info(format, v...)
}

func Fatal(format string, v ...interface{}) {
    Log.Fatal(format, v...)
}