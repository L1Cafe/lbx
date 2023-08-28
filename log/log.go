package log

import (
	"log"
	"strconv"
)

var logLevel uint8

const (
	Info  = uint8(0)
	Warn  = uint8(1)
	Error = uint8(2)
	Fatal = uint8(3)
)

func Init(l uint8) {
	if l < 3 {
		logLevel = l
	} else {
		logLevel = 3
	}
}

func Wrapper(level uint8, msg string) { // TODO make this a variadic function with: client, server, site, response code
	if logLevel == 0 {
		Init(3)
	}
	var levelStr string
	fatal := false
	switch level {
	case Info:
		levelStr = "INFO "
	case Warn:
		levelStr = "WARN "
	case Error:
		levelStr = "ERROR"
	case Fatal:
		levelStr = "FATAL"
		fatal = true
	default:
		levelStr = strconv.Itoa(int(level))
	}
	if fatal {
		log.Fatalf("[%s] %s\n", levelStr, msg)
	} else {
		if level <= logLevel {
			log.Printf("[%s] %s\n", levelStr, msg)
		}
	}
}
