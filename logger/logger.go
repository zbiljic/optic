package logger

import (
	"log"
	"os"
	"strings"
	"sync"
)

const (
	LOGLEVEL_ENV_VAR = "OPTIC_LOGLEVEL"
	LOGFILE_ENV_VAR  = "OPTIC_LOGFILE"

	DEFAULT_LOGLEVEL = LevelInfo
)

var (
	logLevel        Level
	logFile         string
	loggerLock      sync.Mutex
	logLevelFromEnv bool
	logFileFromEnv  bool
	oFile           *os.File
)

// Initialize this logger once
var once sync.Once

func init() {
	once.Do(initLogger)
}

func initLogger() {
	logLevel = DEFAULT_LOGLEVEL

	envLevel := os.Getenv(LOGLEVEL_ENV_VAR)
	if envLevel != "" {
		envLevel = strings.ToUpper(envLevel)
		for level, name := range prefixes {
			if name == envLevel {
				logLevelFromEnv = true
				logLevel = level
				break
			}
		}
	}

	logFile, logFileFromEnv = os.LookupEnv(LOGFILE_ENV_VAR)

	SetupLogging(false, false, "")
}

// SetupLogging configures the logging output.
func SetupLogging(debug, quiet bool, logfile string) {
	loggerLock.Lock()
	defer loggerLock.Unlock()

	filter := NewLevelFilter(DEFAULT_LOGLEVEL)

	// environment variables take precedence
	if logLevelFromEnv {
		filter.SetLevel(logLevel)
	} else {
		if debug {
			filter.SetLevel(LevelDebug)
		}
		if quiet {
			filter.SetLevel(LevelError)
		}
	}
	if logFileFromEnv {
		logfile = logFile
	}

	if logfile != "" {
		if _, err := os.Stat(logfile); os.IsNotExist(err) {
			if oFile, err = os.Create(logfile); err != nil {
				log.Printf("ERROR Unable to create %s (%s), using stderr", logfile, err)
				oFile = os.Stderr
			}
		} else {
			if oFile != nil && oFile != os.Stderr {
				oFile.Close()
			}
			if oFile, err = os.OpenFile(logfile, os.O_APPEND|os.O_WRONLY, os.ModeAppend); err != nil {
				log.Printf("ERROR Unable to append to %s (%s), using stderr", logfile, err)
				oFile = os.Stderr
			}
		}
	} else {
		oFile = os.Stderr
	}

	filter.SetLogOutput(oFile)

	log.SetOutput(filter)
}
