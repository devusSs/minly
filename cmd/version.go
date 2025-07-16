package cmd

import (
	"fmt"

	"github.com/devusSs/minly/internal/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information and exit",
	Run: func(_ *cobra.Command, _ []string) {
		b := version.GetBuild()

		if versionPrintJSON {
			fmt.Println(b.JSON())
			return
		}

		if versionPrintGoString {
			fmt.Println(b.String())
			return
		}

		fmt.Println(b.Pretty())
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
