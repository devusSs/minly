package cmd

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/devusSs/minly/internal/config"
	"github.com/devusSs/minly/internal/log"
	"github.com/devusSs/minly/internal/storage"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var files []storage.File

var filesCmd = &cobra.Command{
	Use:   "files",
	Short: "List, manage and delete files and their links",
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		err := log.Setup()
		checkErr(err, "failed to setup log package")

		go func() {
			err = log.CleanOld()
			if err != nil {
				log.Logger().Error().Err(err).Msg("failed to clean old log files")
			}
		}()

		cfg, err = config.Read()
		logErr(err, "failed to read configuration")
	},
	PersistentPostRun: func(_ *cobra.Command, _ []string) {
		err := log.Flush()
		checkErr(err, "failed to flush log package")
	},
	Run: func(_ *cobra.Command, _ []string) {
		fs, err := storage.NewFileStore()
		logErr(err, "failed to create file store")

		files, err = fs.LoadAll()
		logErr(err, "failed to load files")

		err = printFilesAsTable()
		logErr(err, "failed to print files as table")
	},
}

func init() {
	rootCmd.AddCommand(filesCmd)
}

func printFilesAsTable() error {
	if cfg == nil {
		return errors.New("configuration is not loaded")
	}

	if len(files) == 0 {
		return errors.New("no files found")
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"Timestamp", "Minio Key", "YOURLS Key"})

	for _, f := range files {
		ts := f.Timestamp.UTC().Format(time.RFC3339)

		minURL, err := url.Parse(f.MinioLink)
		if err != nil {
			return fmt.Errorf("failed to parse Minio link: %w", err)
		}

		var yourlsURL *url.URL
		yourlsURL, err = url.Parse(f.YOURLSLink)
		if err != nil {
			return fmt.Errorf("failed to parse YOURLS link: %w", err)
		}

		minioKey := strings.TrimPrefix(minURL.Path, "/")
		yourlsKey := strings.TrimPrefix(yourlsURL.Path, "/")

		err = table.Append([]string{ts, minioKey, yourlsKey})
		if err != nil {
			return fmt.Errorf("failed to append row to table: %w", err)
		}
	}

	return table.Render()
}
