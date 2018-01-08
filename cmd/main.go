package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"sync"

	"github.com/cheggaaa/pb"
	"github.com/spf13/cobra"

	"github.com/zbiljic/optic/pkg/sysinfo"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "optic",
	Short: "Data collection, processing and aggregation server.",
	Long:  `Data collection, processing and aggregation server.`,
}

// Main starts application
func Main() {
	// run exit function only once
	var once sync.Once
	defer once.Do(mainExitFn)

	// Fetch terminal size, if not available, automatically
	// set globalQuiet to true.
	if w, e := pb.GetTerminalWidth(); e != nil {
		globalQuiet = true
	} else {
		globalTermWidth = w
	}

	go func() {
		select {
		case err := <-exitError:
			fmt.Println(err)
			once.Do(mainExitFn)
			exit()
		}
	}()

	if err := rootCmd.Execute(); err != nil {
		exitError <- err
	}
}

var mainExitFn = func() {
	log.Printf("DEBUG Version:%s, Build-time:%s, Commit:%s", Version, BuildTime, ShortCommitID)
	log.Println("DEBUG", sysinfo.GetSysInfo())
}

func init() {
	log.SetOutput(ioutil.Discard)

	defineFlagsGlobal(rootCmd.PersistentFlags())

	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		log.Printf("DEBUG Running command '%s'", cmd.CommandPath())
		registerBefore(cmd)
	}
}

func registerBefore(cmd *cobra.Command) error {

	// Bind global flags to viper config.
	bindFlagsGlobal(cmd)

	// Update global flags (if anything changed from other sources).
	updateGlobals()

	// Configure global flags.
	configureGlobals(cmd)

	return nil
}

// validateCommandCall prints an error if the precondition function returns
// false for a command
func validateCommandCall(cmd *cobra.Command, preconditionFunc func() bool) {
	if !preconditionFunc() {
		cmdName := cmd.Name()
		// prepend parent commands
		tempCmd := cmd
		for {
			if !tempCmd.HasParent() {
				break
			}
			if tempCmd.Parent() == cmd.Root() {
				break
			}
			tempCmd = tempCmd.Parent()
			cmdName = tempCmd.Name() + " " + cmdName
		}
		// The command line is wrong. Print help message and stop.
		fatalIf(errInvalidCommandCall(cmdName), "Invalid command call.")
	}
}
