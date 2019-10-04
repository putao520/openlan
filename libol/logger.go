package libol

import (
	"fmt"
	"log"
	"os"
)

const (
	PRINT = 0x00
	DEUBG = 0x01
	INFO  = 0x02
	WARN  = 0x03
	ERROR = 0x04
	FATAL = 0xff
)

type Logger struct {
	Level		int
	Errors		[]string
	FileName	string
	FileLog		*log.Logger
}

func (l *Logger) Debug(format string, v ...interface{}) {
	if DEUBG >= l.Level {
		log.Printf(fmt.Sprintf("DEBUG %s", format), v...)
	}
}

func (l *Logger) Info(format string, v ...interface{}) {
	if INFO >= l.Level {
		log.Printf(fmt.Sprintf("INFO %s", format), v...)
	}
}

func (l *Logger) Warn(format string, v ...interface{}) {
	if WARN >= l.Level {
		log.Printf(fmt.Sprintf("WARN %s", format), v...)
	}

	l.SaveError(fmt.Sprintf("WARN %s", format), v...)
}

func (l *Logger) Error(format string, v ...interface{}) {
	if ERROR >= l.Level {
		log.Printf(fmt.Sprintf("ERROR %s", format), v...)
	}
	l.SaveError(fmt.Sprintf("ERROR %s", format), v...)
}

func (l *Logger) Fatal(format string, v ...interface{}) {
	if FATAL >= l.Level {
		log.Printf(fmt.Sprintf("FATAL %s", format), v...)
	}

	l.SaveError(fmt.Sprintf("FATAL %s", format), v...)
}

func (l *Logger) Print(format string, v ...interface{}) {
	if PRINT >= l.Level {
		log.Printf(fmt.Sprintf("PRINT %s", format), v...)
	}
}

func (l *Logger) SaveError(format string, v ...interface{}) {
	m := fmt.Sprintf(format, v...)
	if l.FileLog != nil {
		l.FileLog.Println(m)
	}

	if ERROR >= l.Level {
		l.Errors = append(l.Errors, m)
	}
}

var Log = Logger{
	Level:  INFO,
	FileName: ".log.error",
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

func Init (file string, level int) {
	Log.FileName = file
	if Log.FileName != "" {
		logFile, err := os.Create(Log.FileName)
		if err == nil {
			Log.FileLog = log.New(logFile,"", log.LstdFlags)
		} else {
			Warn("logger.Init: %s", err)
		}
	}
}

func SetLog(level int) {
	Log.Level = level
}

func Close() {
	//TODO
}
