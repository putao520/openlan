package libol

import (
    "log"
    "fmt"
)

const (
    PRINT  = 0x00
    DEUBG  = 0x01
    INFO   = 0x02
    ERROR  = 0x03
    FATAL  = 0xff
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
    this.SaveError(fmt.Sprintf("ERROR %s", format), v...)
    log.Printf(fmt.Sprintf("ERROR %s", format), v...)
}

func (this *Logger) Debug(format string, v ...interface{}) {
    log.Printf(fmt.Sprintf("DEBUG %s", format), v...)
}

func (this *Logger) Fatal(format string, v ...interface{}) {
    this.SaveError(fmt.Sprintf("FATAL %s", format), v...)
    log.Printf(fmt.Sprintf("FATAL %s", format), v...)
}

func (this *Logger) Print(format string, v ...interface{}) {
    log.Printf(fmt.Sprintf("PRINT %s", format), v...)
}

func (this *Logger) SaveError(format string, v ...interface{}) {
    ////TODO save to log when too large.
    this.Errors = append(this.Errors, fmt.Sprintf(format, v...))
}

var Log = Logger {
    Level: 1,
    Errors: make([]string, 0, 1024),
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

func Warn(format string, v ...interface{}) {
    Log.Warn(format, v...)
}

func Fatal(format string, v ...interface{}) {
    Log.Fatal(format, v...)
}