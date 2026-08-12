package cmd

import (
	"github.com/spf13/cobra"

	"github.com/devusSs/minly/internal/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information and exit",
	Run: func(cmd *cobra.Command, _ []string) {
		b := version.GetBuild()

		if versionPrintJSON {
			cmd.Println(b.JSON())
			return
		}

		if versionPrintGoString {
			cmd.Println(b.String())
			return
		}

		cmd.Println(b.Pretty())
	},
}

var (
	versionPrintJSON     bool
	versionPrintGoString bool
)

func init() {
	rootCmd.AddCommand(versionCmd)

	versionCmd.Flags().
		BoolVarP(&versionPrintJSON, "json", "j", false, "Print version information in JSON format")
	versionCmd.Flags().
		BoolVarP(&versionPrintGoString, "go-string", "g", false, "Print version information in Go string format")
}
