package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Place for all the flags for all commands

var globalFlags = map[string]func(*pflag.FlagSet){
	"quiet": func(flags *pflag.FlagSet) {
		flags.Bool("quiet", false, "Suppress chatty console output.")
	},
	"debug": func(flags *pflag.FlagSet) {
		flags.Bool("debug", false, "Run the command with debug information in the output.")
	},
	"log-file": func(flags *pflag.FlagSet) {
		flags.String("log-file", "", "File to which to send logs to.")
	},
	"config": func(flags *pflag.FlagSet) {
		flags.String("config", "", "Configuration file to use instead of the default one.")
	},
	"pprof-addr": func(flags *pflag.FlagSet) {
		flags.String("pprof-addr", "", "pprof address to listen on, format: localhost:6060 or :6060.")
	},
}

func defineFlagsGlobal(flags *pflag.FlagSet) {
	for _, ffn := range globalFlags {
		ffn(flags)
	}
	// SPECIAL CASE: Support running as a Windows Service
	if globalOsIsWindows {
		flags.StringVar(&globalServiceCommand, "service", "", "Operate on the service.")
	}
}

func bindFlagsGlobal(cmd *cobra.Command) {
	globalSection := configSections["global"]
	for k := range globalFlags {
		key := globalSection(k)
		log.Printf("TRACE Binding global flag: '%s' to '%s'", k, key)
		viper.BindPFlag(key, cmd.Flag(k))
	}
}
