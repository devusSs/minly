package cmd

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/devusSs/minly/internal/lastrun"
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

		lastRun, err := lastrun.Read()
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				checkErr(err, "failed to read last run data")
			}
		}

		if lastRun.Error != "" {
			_, err = fmt.Fprintf(
				os.Stderr,
				"WARNING: last run error: %s (timestamp: %s)\n",
				lastRun.Error,
				lastRun.Timestamp.Format(time.RFC3339),
			)
			checkErr(err, "failed to write last run error message")
		}
	},
	PersistentPostRun: func(_ *cobra.Command, _ []string) {
		// TODO: catch error from a global variable or function
		err := lastrun.Write(nil)
		checkErr(err, "failed to write last run data")
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
