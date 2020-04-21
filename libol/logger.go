package libol

import (
	"container/list"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"sync"
)

const (
	PRINT = 00
	LOG   = 01
	STACK = 9
	DEBUG = 10
	CMD   = 11
	INFO  = 20
	WARN  = 30
	ERROR = 40
	FATAL = 99
)

var levels = map[int]string{
	PRINT: "PRINT",
	LOG:   "LOG",
	DEBUG: "DEBUG",
	STACK: "STACK",
	CMD:   "CMD",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
	FATAL: "FATAL",
}

type Logger struct {
	Level    int
	FileName string
	FileLog  *log.Logger

	lock   sync.Mutex
	errors *list.List
}

func (l *Logger) Write(level int, format string, v ...interface{}) {
	str, ok := levels[level]
	if !ok {
		str = "NiL"
	}
	if level >= l.Level {
		log.Printf(fmt.Sprintf("%s %s", str, format), v...)
	}
	if level >= INFO {
		l.Save(fmt.Sprintf("%s %s", str, format), v...)
	}
}

func (l *Logger) Save(format string, v ...interface{}) {
	m := fmt.Sprintf(format, v...)
	if l.FileLog != nil {
		l.FileLog.Println(m)
	}
	l.lock.Lock()
	defer l.lock.Unlock()
	if l.errors.Len() >= 1024 {
		if e := l.errors.Front(); e != nil {
			l.errors.Remove(e)
		}
	}
	l.errors.PushBack(m)
}

var logger = Logger{
	Level:    INFO,
	FileName: ".log.error",
	errors:   list.New(),
}

func Print(format string, v ...interface{}) {
	logger.Write(PRINT, format, v...)
}

func Log(format string, v ...interface{}) {
	logger.Write(LOG, format, v...)
}

func Stack(format string, v ...interface{}) {
	logger.Write(STACK, format, v...)
}

func Debug(format string, v ...interface{}) {
	logger.Write(DEBUG, format, v...)
}

func Cmd(format string, v ...interface{}) {
	logger.Write(CMD, format, v...)
}

func Info(format string, v ...interface{}) {
	logger.Write(INFO, format, v...)
}

func Warn(format string, v ...interface{}) {
	logger.Write(WARN, format, v...)
}

func Error(format string, v ...interface{}) {
	logger.Write(ERROR, format, v...)
}

func Fatal(format string, v ...interface{}) {
	logger.Write(FATAL, format, v...)
}

func Init(file string, level int) {
	SetLog(level)
	logger.FileName = file
	if logger.FileName != "" {
		logFile, err := os.Create(logger.FileName)
		if err == nil {
			logger.FileLog = log.New(logFile, "", log.LstdFlags)
		} else {
			Warn("logger.Init: %s", err)
		}
	}
}

func SetLog(level int) {
	logger.Level = level
}

func Close() {
	//TODO
}

func Catch(name string) {
	if err := recover(); err != nil {
		Fatal("%s Panic: >>> %s <<<", name, err)
		Fatal("%s Stack: >>> %s <<<", name, debug.Stack())
	}
}
