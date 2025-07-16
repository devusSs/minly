package cmd

import (
	"fmt"
	"os"

	"github.com/devusSs/minly/internal/system"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "minly",
	Short: "A Go CLI app to combine MinIO and YOURLS.",
	Long: `minly is a Go CLI application that allows you to manage and interact with MinIO and YOURLS.
It provides commands to upload files to MinIO and create short URLs using YOURLS.
It also supports managing and deleting files uploaded to MinIO.

To get started simply run 'minly init' to setup a configuration and needed secrets.

For questions regarding commands simply run 'minly help <command>' or 'minly <command> -h'.

For more help or information check out the GitHub repository.`,
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		checkErr(system.CheckSupported(), "unsupported operating system or architecture")
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func checkErr(err error, msg string) {
	if err != nil {
		cobra.CheckErr(fmt.Sprintf("%s: %v", msg, err))
	}
}
