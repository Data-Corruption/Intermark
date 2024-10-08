package utils

import (
	"log"
	"os"
	"time"

	"github.com/Data-Corruption/blog"
)

const logPath = "logs"

var (
	initialized bool = false
	DebugMode   bool
)

func InitLogger() {
	// Get log level from config
	logLevel, err := blog.LogLevelFromString(Config.LogLevel)
	if err != nil {
		log.Fatalf("'%s' is not a valid log level", Config.LogLevel)
	}
	// Create the log directory if it doesn't exist
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		if err := os.MkdirAll(logPath, os.ModePerm); err != nil {
			log.Fatalf("Error creating log directory: %s", err)
		}
	}
	// Initialize the logger
	if err := blog.Init(logPath, logLevel); err != nil {
		log.Fatalf("Error initializing logger: %s", err)
	}
	DebugMode = false
	if Config.LogLevel == "debug" {
		blog.SetUseConsole(true)
		DebugMode = true
	}
	initialized = true
	blog.Info("Logger initialized")
}

func CleanupLogger() {
	if initialized {
		time.Sleep(1 * time.Second)
		blog.SyncFlush(0)
	}
}
