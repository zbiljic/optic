package cmd

import (
	"time"

	"github.com/spf13/cobra"
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Gather events once, print them to stdout, and exit",
	Long:  `Gather events once, print them to stdout, and exit`,
	PreRun: func(cmd *cobra.Command, args []string) {
		validateCommandCall(cmd, func() bool {
			return len(args) == 0
		})
	},
	Run: func(cmd *cobra.Command, args []string) {
		defer timeTrack(time.Now(), "test")
		testMain()
	},
}

func init() {
	// add 'test' command to root command
	rootCmd.AddCommand(testCmd)
}

func testMain() error {
	// test command is run
	globalTestCommand = true

	return agentMain()
}
