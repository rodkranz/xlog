package xlog

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"path/filepath"
)

const version = "0.0.1"

func Version() string {
	return version
}

type (
	MODE string
	LEVEL int
)

const (
	TRACE LEVEL = iota
	INFO
	WARN
	ERROR
	FATAL
)

type Fields map[string]interface{}

var formats = map[LEVEL]string{
	TRACE: "[TRACE] ",
	INFO:  "[ INFO] ",
	WARN:  "[ WARN] ",
	ERROR: "[ERROR] ",
	FATAL: "[FATAL] ",
}

func isValidLevel(level LEVEL) bool {
	return level >= TRACE && level <= FATAL
}

type Message struct {
	Level LEVEL
	Body  []interface{}
}

func Write(level LEVEL, skip int, v ...interface{}) {
	msg := &Message{
		Level: level,
		Body:  v,
	}

	if msg.Level >= ERROR && skip > 0 {
		pc, file, line, ok := runtime.Caller(skip)
		if ok {
			// Get caller function name.
			fn := runtime.FuncForPC(pc)
			var fnName string
			if fn == nil {
				fnName = "?()"
			} else {
				fnName = strings.TrimLeft(filepath.Ext(fn.Name()), ".") + "()"
			}

			if len(file) > 20 {
				file = "..." + file[len(file)-20:]
			}
			msg.Body = append(msg.Body, formats[level]+fmt.Sprintf("[%s:%d %s] ", file, line, fnName)+fmt.Sprint(v...))
		}
	}

	for i := range receivers {
		if receivers[i].Level() > level {
			continue
		}

		receivers[i].msgChan <- msg
	}
}

func Trace(v ...interface{}) {
	Write(TRACE, 0, v...)
}

func Info(v ...interface{}) {
	Write(INFO, 0, v...)
}

func Warn(v ...interface{}) {
	Write(WARN, 0, v...)
}

func Error(skip int, v ...interface{}) {
	Write(ERROR, skip, v...)
}

func Fatal(skip int, v ...interface{}) {
	Write(FATAL, skip, v...)
	Shutdown()
	os.Exit(1)
}

func Shutdown() {
	for i := range receivers {
		receivers[i].Destroy()
	}

	// Shutdown the error handling goroutine.
	quitChan <- struct{}{}
	for {
		if len(errorChan) == 0 {
			break
		}

		fmt.Printf("xlog: unable to write message: %v\n", <-errorChan)
	}
}
