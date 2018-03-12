package xlog

import (
	"fmt"
	"os"
)

const version = "0.0.1"

func Version() string {
	return version
}

type (
	MODE  string
	LEVEL int
)

const (
	TRACE LEVEL = iota
	INFO
	WARN
	ERROR
	FATAL
)

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

func writer(level LEVEL, skip int, v ...interface{}) {
	msg := &Message{
		Level: level,
		Body:  v,
	}

	for i := range receivers {
		if receivers[i].Level() > level {
			continue
		}

		receivers[i].msgChan <- msg
	}
}

func Trace(v ...interface{}) {
	writer(TRACE, 0, v...)
}

func Info(v ...interface{}) {
	writer(INFO, 0, v...)
}

func Warn(v ...interface{}) {
	writer(WARN, 0, v...)
}

func Error(skip int, v ...interface{}) {
	writer(ERROR, 0, v...)
}

func Fatal(skip int, v ...interface{}) {
	writer(FATAL, 0, v...)
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
