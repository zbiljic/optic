package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/zbiljic/optic/pkg/console"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version of application currently used",
	Long:  `Display version of application currently used`,
	PreRun: func(cmd *cobra.Command, args []string) {
		validateCommandCall(cmd, func() bool {
			return len(args) == 0
		})
	},
	Run: func(cmd *cobra.Command, args []string) {
		printVersion()
	},
}

func init() {
	// add 'version' command to root command
	rootCmd.AddCommand(versionCmd)
}

// Structured message depending on the type of console.
type versionMessage struct {
	Status    string `json:"status"`
	Version   string `json:"version"`
	BuildTime string `json:"buildTime"`
	CommitID  string `json:"commitID"`
}

// Message for console printing.
func (v versionMessage) String() string {
	str := fmt.Sprintf("optic version %s", v.Version)
	if globalDebug {
		str = str + "\n" +
			fmt.Sprintf("Build-time:   %s\n", v.BuildTime) +
			fmt.Sprintf("Commit-id:    %s", v.CommitID)
	}
	return str
}

func printVersion() error {

	msg := versionMessage{}
	msg.Version = Version
	msg.BuildTime = BuildTime
	msg.CommitID = CommitID

	if !globalQuiet {
		console.Println(msg.String())
	}
	return nil
}

// Check the interfaces are satisfied
var (
	_ fmt.Stringer = &versionMessage{}
)
