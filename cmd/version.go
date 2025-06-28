package cmd

import (
	"fmt"

	"github.com/devusSs/minly/internal/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints version information and exits",
	Long: `Prints version / build information about the application.
It will default to JSON format if nothing is specified, but can also print in a pretty format or plain text format.
Only one output format can be selected at a time (JSON->Pretty->Text).`,
	Run: func(_ *cobra.Command, _ []string) {
		build := version.GetBuild()

		if versionPrintJSON {
			fmt.Println(build.JSON())
			return
		}

		if versionPrintPretty {
			build.PrettyPrint()
			return
		}

		if versionPrintText {
			fmt.Println(build.String())
			return
		}

		fmt.Println(build.JSON())
	},
}

var (
	versionPrintJSON   bool
	versionPrintPretty bool
	versionPrintText   bool
)

func init() {
	rootCmd.AddCommand(versionCmd)

	versionCmd.Flags().
		BoolVarP(&versionPrintJSON, "json", "j", false, "Print version information in JSON format")
	versionCmd.Flags().
		BoolVarP(&versionPrintPretty, "pretty", "p", false, "Print version information in a human-readable format")
	versionCmd.Flags().
		BoolVarP(&versionPrintText, "text", "t", false, "Print version information in plain text format")

}
