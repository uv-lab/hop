package main

import (
	"flag"
	"fmt"
	"os"
)

const (
	Version = "0.1.0"
)

// Define a type to store command options
type Flags struct {
	RedisHost        string
	RedisPort        uint16
	RedisDB          int64
	RedisPWD         string
	RedisKey         string
	RetryTimes       uint8
	RetryInterval    uint16
	WorkerPoolSize   uint16
	LogQueueSize     uint32
	LogBufferSize    uint16
	AdminPort        uint16
	TaskFetchTimeout uint16
	LogLevel         LogLevel
}

var (
	// the global config
	conf *Flags

	// the global task queue
	taskQueue *TaskQueue

	// the global worker pool
	workerPool chan bool

	// a channel use to send shutdown start message
	shutdownStartChan = make(chan bool)

	// a channel use to send shutdown complete message
	shutdownCompChan = make(chan bool)
)

func main() {
	parseFlags()
	logStart()
	statsStart()
	adminStart()
	taskQueue = NewTaskQueue()
	workerHub.run()
	workerPool = NewWorkerPool()

	for {
		Log(LogLevelInfo, "main loop")
		if IsShutdown() {
			Log(LogLevelInfo, "shutdown...")
			// ready for shutdown
			shutdownStartChan <- true
			if stats.WorkerCurr > 0 {
				// wait for workers shutdown
				<-shutdownCompChan
			}
			Log(LogLevelInfo, "shutdown done")
			FlushLog()
			Log(LogLevelInfo, "flush log done")
			return
		}
		task, err := NewTask()
		if err != nil {
			continue
		}
		w := NewWorker(task)
		w.Run()
	}
}

func parseFlags() {
	redisHost := flag.String("redis-host", "127.0.0.1", "redis host")
	redisPort := flag.Uint("redis-port", 6379, "redis port")
	redisDB := flag.Uint("redis-db", 0, "redis db")
	redisPWD := flag.String("redis-pwd", "", "redis password")
	redisKey := flag.String("redis-key", "", "redis key")
	retryTimes := flag.Uint("t", 5, "retry times")
	retryInterval := flag.Uint("i", 10, "retry interval, unit: Second")
	workerPoolSize := flag.Uint("n", 1000, "max pool size of workers")
	logQueueSize := flag.Uint("log-queue", 1000, "log queue size")
	logBufferSize := flag.Uint("log-buffer", 2, "log buffer size")
	adminPort := flag.Uint("admin-port", 8888, "admin port")
	fetchTaskTimeout := flag.Uint("fetch-timeout", 10, "timeout to fetch a task")
	v := flag.Bool("v", true, "log level 1")
	vv := flag.Bool("vv", false, "log level 2")
	vvv := flag.Bool("vvv", false, "log level 3")

	flag.Parse()

	if *redisKey == "" {
		fmt.Println("please specify the redis key")
		os.Exit(1)
	}

	// default log level is 1
	if !*v && !*vv && !*vvv {
		*v = true
	}

	var logLevel LogLevel
	if *v {
		logLevel = LogLevelWarning
	}
	if *vv {
		logLevel = LogLevelInfo
	}
	if *vvv {
		logLevel = LogLevelDebug
	}

	conf = &Flags{
		RedisHost:        *redisHost,
		RedisPort:        uint16(*redisPort),
		RedisDB:          int64(*redisDB),
		RedisPWD:         *redisPWD,
		RedisKey:         *redisKey,
		RetryTimes:       uint8(*retryTimes),
		RetryInterval:    uint16(*retryInterval),
		WorkerPoolSize:   uint16(*workerPoolSize),
		LogQueueSize:     uint32(*logQueueSize),
		LogBufferSize:    uint16(*logBufferSize),
		AdminPort:        uint16(*adminPort),
		TaskFetchTimeout: uint16(*fetchTaskTimeout),
		LogLevel:         logLevel,
	}
	fmt.Println(conf)
}
