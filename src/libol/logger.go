package libol

import (
	"container/list"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

const (
	PRINT = 00
	LOCK  = 01
	LOG   = 05
	STACK = 06
	DEBUG = 10
	CMD   = 15
	CMD1  = 16
	INFO  = 20
	WARN  = 30
	ERROR = 40
	FATAL = 99
)

type Message struct {
	Level   string `json:"level"`
	Date    string `json:"date"`
	Message string `json:"message"`
}

var levels = map[int]string{
	PRINT: "PRINT",
	LOG:   "LOG",
	LOCK:  "LOCK",
	DEBUG: "DEBUG",
	STACK: "STACK",
	CMD:   "CMD",
	CMD1:  "CMD1",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
	FATAL: "FATAL",
}

type _Log struct {
	Level    int
	FileName string
	FileLog  *log.Logger
	Lock     sync.Mutex
	Errors   *list.List
}

func (l *_Log) Write(level int, format string, v ...interface{}) {
	str, ok := levels[level]
	if !ok {
		str = "NULL"
	}
	if level >= l.Level {
		log.Printf(fmt.Sprintf("[%s] %s", str, format), v...)
	}
	if level >= INFO {
		l.Save(str, format, v...)
	}
}

func (l *_Log) Save(level string, format string, v ...interface{}) {
	m := fmt.Sprintf(format, v...)
	if l.FileLog != nil {
		l.FileLog.Println("["+ level + "] " + m)
	}
	l.Lock.Lock()
	defer l.Lock.Unlock()
	if l.Errors.Len() >= 1024 {
		if e := l.Errors.Back(); e != nil {
			l.Errors.Remove(e)
		}
	}
	yy, mm, dd := time.Now().Date()
	hh, mn, se := time.Now().Clock()
	ele := &Message{
		Level:   level,
		Date:    fmt.Sprintf("%d/%02d/%02d %02d:%02d:%02d", yy, mm, dd, hh, mn, se),
		Message: m,
	}
	l.Errors.PushBack(ele)
}

func (l *_Log) List() <-chan *Message {
	c := make(chan *Message, 128)
	go func() {
		l.Lock.Lock()
		defer l.Lock.Unlock()
		for ele := l.Errors.Back(); ele != nil; ele = ele.Prev() {
			c <- ele.Value.(*Message)
		}
		c <- nil // Finish channel by nil.
	}()
	return c
}

var Logger = _Log{
	Level:    INFO,
	FileName: ".log.error",
	Errors:   list.New(),
}

func HasLog(level int) bool {
	if level >= Logger.Level {
		return true
	}
	return false
}

func Print(format string, v ...interface{}) {
	Logger.Write(PRINT, format, v...)
}

func Log(format string, v ...interface{}) {
	Logger.Write(LOG, format, v...)
}

func Lock(format string, v ...interface{}) {
	Logger.Write(LOCK, format, v...)
}

func Stack(format string, v ...interface{}) {
	Logger.Write(STACK, format, v...)
}

func Debug(format string, v ...interface{}) {
	Logger.Write(DEBUG, format, v...)
}

func Cmd(format string, v ...interface{}) {
	Logger.Write(CMD, format, v...)
}

func Cmd1(format string, v ...interface{}) {
	Logger.Write(CMD1, format, v...)
}

func Info(format string, v ...interface{}) {
	Logger.Write(INFO, format, v...)
}

func Warn(format string, v ...interface{}) {
	Logger.Write(WARN, format, v...)
}

func Error(format string, v ...interface{}) {
	Logger.Write(ERROR, format, v...)
}

func Fatal(format string, v ...interface{}) {
	Logger.Write(FATAL, format, v...)
}

func Init(file string, level int) {
	SetLog(level)
	Logger.FileName = file
	if Logger.FileName != "" {
		logFile, err := os.Create(Logger.FileName)
		if err == nil {
			Logger.FileLog = log.New(logFile, "", log.LstdFlags)
		} else {
			Warn("Logger.Init: %s", err)
		}
	}
}

func SetLog(level int) {
	Logger.Level = level
}

func Close() {
	//TODO
}

func Catch(name string) {
	if err := recover(); err != nil {
		Fatal("%s [PANIC] >>> %s <<<", name, err)
		Fatal("%s [STACK] >>> %s <<<", name, debug.Stack())
	}
}
