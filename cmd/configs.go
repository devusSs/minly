package cmd

import (
	"fmt"
	"net/url"
	"os"
	"text/tabwriter"
	"time"

	"github.com/devusSs/minly/internal/config"
	"github.com/spf13/cobra"
)

var configs []*config.Config

var configsCmd = &cobra.Command{
	Use:   "configs",
	Short: "List and delete saved configs",
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		var err error
		configs, err = config.GetSaved()
		cobra.CheckErr(err)
	},
}

var configsListCmdDetailed bool

var configsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all saved configs",
	Run: func(_ *cobra.Command, _ []string) {
		if configsListCmdDetailed {
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

			fmt.Fprintln(
				w,
				"Project Name\tCreated At\tUpdated At\tLogs Dir\tStorages Dir\tSecrets Service\tMinio Endpoint\tMinio Public Bucket\tMinio Private Bucket\tYOURLS Endpoint\tYOURLS Description",
			)
			for _, cfg := range configs {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
					cfg.ProjectName,
					cfg.CreatedAt.Format(time.RFC3339),
					cfg.UpdatedAt.Format(time.RFC3339),
					cfg.LogsDirectory,
					cfg.StoragesDirectory,
					cfg.SecretsServiceName,
					formatURL(cfg.MinioEndpoint),
					cfg.MinioPublicBucketName,
					cfg.MinioPrivateBucketName,
					formatURL(cfg.YOURLSEndpoint),
					cfg.YOURLSDescription,
				)
			}

			err := w.Flush()
			cobra.CheckErr(err)

			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

		fmt.Fprintln(w, "Project Name\tCreated At\tUpdated At\tMinio Endpoint\tYOURLS Endpoint")
		for _, cfg := range configs {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				cfg.ProjectName,
				cfg.CreatedAt.Format(time.RFC3339),
				cfg.UpdatedAt.Format(time.RFC3339),
				formatURL(cfg.MinioEndpoint),
				formatURL(cfg.YOURLSEndpoint),
			)
		}

		err := w.Flush()
		cobra.CheckErr(err)
	},
}

func init() {
	rootCmd.AddCommand(configsCmd)

	configsCmd.AddCommand(configsListCmd)
	configsListCmd.Flags().
		BoolVarP(&configsListCmdDetailed, "detailed", "d", false, "Shows every field in the config")
}

func formatURL(u *url.URL) string {
	if u == nil {
		return "-"
	}
	return u.String()
}
