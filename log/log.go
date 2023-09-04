package log

import (
	"log"
	"strconv"
)

var logLevel uint8

const (
	Info  = uint8(1)
	Warn  = uint8(2)
	Error = uint8(3)
	Fatal = uint8(4)
)

func Init(l uint8) {
	if l < 4 {
		logLevel = l
	} else {
		logLevel = 4
		log.Printf("Log level not specified, defaulting to level 4")
	}
}

func Wrapper(level uint8, msg string) { // TODO make this a variadic function with: client, server, site, response code
	if logLevel == 0 {
		// TODO check that when not setting the log level, the application starts with level 4
		Init(4)
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
