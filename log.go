package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
)

const (
	LevelDebug   = "DEBUG"
	LevelInfo    = "INFO"
	LevelNotice  = "NOTICE"
	LevelWarning = "WARNING"
	LevelError   = "ERROR"
)

type logData struct {
	message string
	level   string
}

var (
	logChannel     chan *logData
	buf            bytes.Buffer
	logger         *log.Logger
	currBufferSize uint16 = 0
)

func logStart() {
	logChannel = make(chan *logData, flags.LogQueueSize)
	logger = log.New(&buf, "", log.LstdFlags)
	go func() {
		for {
			select {
			case aLog := <-logChannel:
				writeLog(aLog)
			}
		}
	}()
}

func Logf(level string, format string, a ...interface{}) {
	message := fmt.Sprintf(format, a...)
	aLog := &logData{message, level}
	logChannel <- aLog
}

func Log(level string, a ...interface{}) {
	message := fmt.Sprint(a...)
	aLog := &logData{message, level}
	logChannel <- aLog
}

func writeLog(aLog *logData) {
	logger.Printf("[%s] %s\n", aLog.level, aLog.message)
	if aLog.level == LevelError {
		defer func() { os.Exit(1) }()
		logger.Printf("SHUTDOWN\n")
	}
	currBufferSize++
	if currBufferSize >= flags.LogBufferSize {
		log.Print(&buf)
		buf.Reset()
		currBufferSize = 0
	}
}
