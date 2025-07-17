package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/devusSs/minly/internal/config"
	"github.com/devusSs/minly/internal/log"
	"github.com/devusSs/minly/internal/secret"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show, edit or delete the configuration and secrets",
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		err := log.Setup()
		checkErr(err, "failed to setup log package")

		go func() {
			err = log.CleanOld()
			if err != nil {
				log.Logger().Error().Err(err).Msg("failed to clean old log files")
			}
		}()
	},
	PersistentPostRun: func(_ *cobra.Command, _ []string) {
		err := log.Flush()
		checkErr(err, "failed to flush log package")
	},
	Run: func(_ *cobra.Command, _ []string) {
		var err error
		cfg, err = config.Read()
		logErr(err, "failed to read config")

		log.Logger().Debug().Msg("configuration loaded successfully")

		var minioAccessKey, minioAccessSecret, yourlsSignature string

		if configShowSensitive {
			minioAccessKey, err = getSecret(secret.MinioAccessKey)
			logErr(err, "failed to get MinIO access key")

			log.Logger().Debug().Msg("got MinIO access key")

			minioAccessSecret, err = getSecret(secret.MinioAccessSecret)
			logErr(err, "failed to get MinIO access secret")

			log.Logger().Debug().Msg("got MinIO access secret")

			yourlsSignature, err = getSecret(secret.YOURLSignature)
			logErr(err, "failed to get YOURLS signature")

			log.Logger().Debug().Msg("got YOURLS signature")
		}

		fmt.Println("Configuration")
		fmt.Println("-------------")
		fmt.Printf("Project Name:\t\t%s\n", cfg.ProjectName)
		fmt.Printf("Created At:\t\t%s\n", cfg.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated At:\t\t%s\n", cfg.UpdatedAt.Format(time.RFC3339))
		fmt.Printf("MinIO Endpoint:\t\t%s\n", cfg.MinioEndpoint)
		fmt.Printf("MinIO Use SSL:\t\t%t\n", cfg.MinioUseSSL)
		fmt.Printf("MinIO Bucket Name:\t%s\n", cfg.MinioBucketName)
		fmt.Printf("MinIO Region:\t\t%s\n", cfg.MinioRegion)
		fmt.Printf("MinIO Link Expiry:\t%s\n", cfg.MinioLinkExpiry.String())
		fmt.Printf("YOURLS Endpoint:\t%s\n", cfg.YOURLSEndpoint)

		if configShowSensitive {
			log.Logger().Debug().Msg("printing sensitive information")

			fmt.Println()
			fmt.Println("Secrets")
			fmt.Println("-------")
			fmt.Printf("MinIO Access Key:\t%s\n", minioAccessKey)
			fmt.Printf("MinIO Access Secret:\t%s\n", minioAccessSecret)
			fmt.Printf("YOURLS Signature:\t%s\n", yourlsSignature)
		}
	},
}

var configShowSensitive bool

var configDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete the configuration and optionally the secrets",
	Run: func(_ *cobra.Command, _ []string) {
		var err error
		cfg, err = config.Read()
		logErr(err, "failed to read config")

		log.Logger().Debug().Msg("configuration loaded successfully")

		if configDeleteSecrets {
			err = secret.DeleteAll()
			logErr(err, "failed to delete secrets")

			log.Logger().Info().Msg("secrets deleted successfully")
		}

		err = os.Remove(cfg.FilePath())
		logErr(err, "failed to delete config file")

		log.Logger().Info().Msg("configuration deleted successfully")
	},
}

var configDeleteSecrets bool

func init() {
	rootCmd.AddCommand(configCmd)

	configCmd.Flags().
		BoolVar(&configShowSensitive, "show-sensitive", false, "Show secrets as well as the configuration")

	configCmd.AddCommand(configDeleteCmd)

	configDeleteCmd.Flags().
		BoolVar(&configDeleteSecrets, "secrets", false, "Delete secrets in addition to the configuration")
}

func getSecret(key secret.Key) (string, error) {
	value, err := secret.Load(key)
	if err != nil {
		return "", fmt.Errorf("failed to load secret %s: %w", key, err)
	}

	return value, nil
}
