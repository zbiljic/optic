package cmd

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/zbiljic/optic/pkg/console"
	"github.com/zbiljic/optic/pkg/sysinfo"
)

var (
	// Error exit status code that will be used in `Main` if the application
	// returned an error.
	errorExitStatusCode = globalErrorExitStatus

	// Channel for errors upon which application will exit.
	exitError = make(chan error, 1)
)

// fatalIf wrapper function which takes error and selectively prints stack
// frames if available on debug
func fatalIf(err error, msg string) {
	if err == nil {
		return
	}
	log.Println("ERROR", msg, fmt.Sprintf("%+v", err))

	if !globalDebug {
		console.Fatalln(fmt.Sprintf("%s %s", msg, err.Error()))
	}
	sysInfo := sysinfo.GetSysInfo()
	console.Fatalln(fmt.Sprintf("%s %+v", msg, err), "\n", sysInfo)
}

// errorIf synonymous with fatalIf but doesn't exit on error != nil
func errorIf(err error, msg string) {
	if err == nil {
		return
	}
	log.Println("ERROR", msg, fmt.Sprintf("%+v", err))

	if !globalDebug {
		console.Errorln(fmt.Sprintf("%s %s", msg, err.Error()))
		return
	}
	sysInfo := sysinfo.GetSysInfo()
	console.Errorln(fmt.Sprintf("%s %+v", msg, err), "\n", sysInfo)
}

// exitStatus allows setting custom exitStatus number. It returns empty error.
func exitStatus(status int) error {
	errorExitStatusCode = status
	err := errDummy()
	exitError <- err
	return err
}

func exit() {
	var wg sync.WaitGroup
	wg.Add(1)
	os.Exit(errorExitStatusCode)
	wg.Wait()
}
