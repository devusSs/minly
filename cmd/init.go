package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/devusSs/minly/internal/config"
	"github.com/devusSs/minly/internal/log"
	"github.com/devusSs/minly/internal/secret"
	"github.com/devusSs/minly/internal/version"
	"github.com/spf13/cobra"
)

var cfg *config.Config

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes the config and needed secrets",
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
		log.Logger().Debug().Any("build", version.GetBuild()).Msg("init")

		var err error
		cfg, err = config.Read()
		if err == nil && initOverwrite {
			logErr(
				errors.New("config exists"),
				"config already exists, use --overwrite to overwrite it",
			)
		}

		log.Logger().Debug().Msg("no previous config found, initializing")

		switch {
		case initUseFile:
			if initFilePath == "" {
				logErr(
					errors.New("file path not provided"),
					"file path must be provided when using --file",
				)
			}

			config.SetFile(initFilePath)
			cfg, err = config.FromFile()
			logErr(err, "failed to read config from file")

			log.Logger().Info().Str("file", initFilePath).
				Msg("config initialized from file")

		case initUseEnv:
			config.SetEnvFile(initEnvFilePath)
			cfg, err = config.FromEnv()
			logErr(err, "failed to read config from environment variables")

			log.Logger().Info().Str("env-file", initEnvFilePath).
				Msg("config initialized from environment variables")

		default:
			cfg, err = config.FromInput()
			logErr(err, "failed to read config from input")

			fmt.Println()
			log.Logger().Info().Msg("config initialized from input")
		}

		err = checkOrSetSecret(secret.MinioAccessKey, initReSetSecrets)
		logErr(err, "failed to check or set MinIO access key")

		log.Logger().Info().Msg("MinIO access key set")

		err = checkOrSetSecret(secret.MinioAccessSecret, initReSetSecrets)
		logErr(err, "failed to check or set MinIO access secret")

		log.Logger().Info().Msg("MinIO access secret set")

		err = checkOrSetSecret(secret.YOURLSignature, initReSetSecrets)
		logErr(err, "failed to check or set YOURLS signature")

		log.Logger().Info().Msg("YOURLS signature set")

		log.Logger().Info().Msg("secrets initialized successfully")

		err = config.Write(cfg)
		logErr(err, "failed to write config")

		log.Logger().Info().Msg("config written successfully")
		log.Logger().Info().Msg("minly initialized successfully")
	},
}

var (
	initUseFile      bool
	initFilePath     string
	initUseEnv       bool
	initEnvFilePath  string
	initOverwrite    bool
	initReSetSecrets bool
)

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().BoolVar(&initUseFile, "file", false, "Initialize config from file")
	initCmd.Flags().StringVar(&initFilePath, "file-path", "", "Path to the config file")
	initCmd.Flags().
		BoolVar(&initUseEnv, "env", false, "Initialize config from environment variables")
	initCmd.Flags().
		StringVar(&initEnvFilePath, "env-file-path", "", "Path to the environment variables file if desired")
	initCmd.Flags().
		BoolVar(&initOverwrite, "overwrite", false, "Overwrite existing config")
	initCmd.Flags().
		BoolVar(&initReSetSecrets, "reset-secrets", false, "Re-set secrets on keychain")

	initCmd.MarkFlagsMutuallyExclusive("file", "env")
	initCmd.MarkFlagsRequiredTogether("file", "file-path")
}

func logErr(err error, msg string) {
	if err != nil {
		log.Logger().Error().Err(err).Msg(msg)
		os.Exit(1)
	}
}

func checkOrSetSecret(key secret.Key, reset bool) error {
	exists, err := secret.Exists(key)
	if err != nil {
		return fmt.Errorf("failed to check secret %s: %w", key, err)
	}

	if !exists || reset {
		var input string
		input, err = secret.GetInput(fmt.Sprintf("Enter a value for %s", key))
		if err != nil {
			return fmt.Errorf("failed to get input for secret %s: %w", key, err)
		}

		err = secret.Save(key, input)
		if err != nil {
			return fmt.Errorf("failed to save secret %s: %w", key, err)
		}

		return nil
	}

	return nil
}
