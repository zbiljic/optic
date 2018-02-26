package cmd

import (
	"runtime"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/zbiljic/optic/pkg/console"
)

const (
	globalErrorExitStatus = 1 // Global error exit status.
)

var (
	globalQuiet     = false // Quiet flag set via command line
	globalDebug     = false // Debug flag set via command line
	globalLogFile   = ""    // Logfile flag set via command line
	globalConfig    = ""    // Config flag set via command line
	globalPprofAddr = ""    // pprof address flag set via command line
	// WHEN YOU ADD NEXT GLOBAL FLAG, MAKE SURE TO ALSO UPDATE PERSISTENT FLAGS, FLAG CONSTANTS AND UPDATE FUNC.
)

// Special variables
var (
	globalConfigFile = false // If config file has been explicitly set

	globalOsIsWindows    = false // If the application is started on Windows
	globalServiceCommand = ""

	globalTestCommand = false // If the test command is called
)

var (
	// Terminal width
	globalTermWidth int
)

func init() {
	if runtime.GOOS == "windows" {
		globalOsIsWindows = true
	}
}

var (
	configSections = map[string]func(string) string{
		"global": nil,
	}
	// WHEN YOU ADD NEXT GLOBAL CONFIG SECTION, MAKE SURE TO ALSO UPDATE TESTS, ETC.
)

var _ error = initConfigSections()

func initConfigSections() error {
	// DO NOT EDIT - builds section functions
	for k := range configSections {
		localK := k
		configSections[localK] = func(key string) string {
			return localK + "." + key
		}
	}
	return nil
}

func configureGlobals(cmd *cobra.Command) {
	// Enable debug messages if requested.
	if globalDebug {
		console.DebugPrint = true
	}

	if cmd.Flags().Changed("config") {
		globalConfigFile = true
	}
}

func updateGlobals() {
	globalSection := configSections["global"]
	globalQuiet = viper.GetBool(globalSection("quiet"))
	globalDebug = viper.GetBool(globalSection("debug"))
	globalLogFile = viper.GetString(globalSection("log-file"))
	globalConfig = viper.GetString(globalSection("config"))
	globalPprofAddr = viper.GetString(globalSection("pprof-addr"))
}
