package cmd

import (
	"fmt"
	"strings"

	"github.com/devusSs/minly/internal/config"
	"github.com/devusSs/minly/internal/input"
	"github.com/devusSs/minly/internal/secrets"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Adds a new project if not already present",
	Run: func(_ *cobra.Command, _ []string) {
		configs, err := config.GetSaved()
		cobra.CheckErr(err)

		var projectName string
		projectName, err = input.ReadFromStdin("Enter the project name:")
		cobra.CheckErr(err)

		projectName = strings.ToLower(projectName)

		for _, config := range configs {
			if config.ProjectName == projectName {
				cobra.CheckErr(fmt.Sprintf("Project '%s' already exists", projectName))
			}
		}

		fmt.Printf("Creating a new project called '%s'...\n", projectName)

		fmt.Println()

		var minioEndpoint string
		minioEndpoint, err = input.ReadFromStdin(
			"Enter the Minio endpoint (e.g., https://example.com):",
		)
		cobra.CheckErr(err)

		var yourlsEndpoint string
		yourlsEndpoint, err = input.ReadFromStdin(
			"Enter the YOURLS endpoint (e.g., https://example.com/yourls-api.php):",
		)
		cobra.CheckErr(err)

		fmt.Println()
		fmt.Println("The tool will now prompt you for secrets related to the project.")
		fmt.Println("They will be stored securely in your system's keyring.")

		var cfg *config.Config
		cfg, err = config.NewConfig(projectName, minioEndpoint, yourlsEndpoint)
		cobra.CheckErr(err)

		fmt.Println()

		err = getAndInsertSecrets(cfg.SecretsServiceName)
		cobra.CheckErr(err)

		err = config.WriteConfig(cfg)
		cobra.CheckErr(err)

		fmt.Println()
		fmt.Println("Config saved successfully.")

		fmt.Println()
		fmt.Printf("Project '%s' created successfully.\n", projectName)
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
}

func getAndInsertSecrets(secretsServiceName string) error {
	created, err := ifNotExistsAskAndInsertKey(secretsServiceName, secrets.MinioAccessKeyKey)
	if err != nil {
		return fmt.Errorf("failed to insert Minio Access Key: %w", err)
	}

	if !created {
		fmt.Printf("%s secret already exists, skipping...\n", secrets.MinioAccessKeyKey)
	}

	created, err = ifNotExistsAskAndInsertKey(secretsServiceName, secrets.MinioAccessSecretKey)
	if err != nil {
		return fmt.Errorf("failed to insert Minio Access Secret: %w", err)
	}

	if !created {
		fmt.Printf("%s secret already exists, skipping...\n", secrets.MinioAccessSecretKey)
	}

	created, err = ifNotExistsAskAndInsertKey(secretsServiceName, secrets.YOURLSSignatureKey)
	if err != nil {
		return fmt.Errorf("failed to insert YOURLS Signature Key: %w", err)
	}

	if !created {
		fmt.Printf("%s secret already exists, skipping...\n", secrets.YOURLSSignatureKey)
	}

	return nil
}

func ifNotExistsAskAndInsertKey(serviceName string, key string) (bool, error) {
	exists, err := secrets.Exists(serviceName, key)
	if err != nil {
		return false, fmt.Errorf("failed to check if secret exists: %w", err)
	}

	if exists {
		return false, nil
	}

	value, err := input.ReadFromStdin(fmt.Sprintf("Enter the value for '%s':", key))
	if err != nil {
		return false, fmt.Errorf("failed to read value for '%s': %w", key, err)
	}

	err = secrets.Save(serviceName, key, value)
	if err != nil {
		return false, fmt.Errorf("failed to insert secret '%s': %w", key, err)
	}

	return true, nil
}
