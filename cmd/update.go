package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/devusSs/minly/internal/log"
	"github.com/devusSs/minly/internal/update"
	"github.com/devusSs/minly/internal/version"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "A brief description of your command",
	PreRun: func(_ *cobra.Command, _ []string) {
		err := log.Setup()
		checkErr(err, "failed to setup log package")

		go func() {
			err = log.CleanOld()
			if err != nil {
				log.Logger().Error().Err(err).Msg("failed to clean old log files")
			}
		}()
	},
	PostRun: func(_ *cobra.Command, _ []string) {
		err := log.Flush()
		checkErr(err, "failed to flush log package")
	},
	Run: func(_ *cobra.Command, _ []string) {
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		defer cancel()

		log.Logger().Info().Msg("starting update process")

		u, err := update.DoUpdate(ctx, version.GetBuild().Version)
		logErr(err, "failed to perform update process")

		log.Logger().Debug().Any("update", u).Msg("update process completed")

		if u.Updated {
			log.Logger().
				Info().
				Str("version", u.Version).
				Time("date", u.Date).
				Msg("update completed successfully, please restart the application to apply changes")

			fmt.Println()
			fmt.Println("Changelog:")
			fmt.Println(u.Changelog)
			return
		}

		log.Logger().Info().Msg("no updates available")
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
