package cmd

import (
	"context"
	"errors"
	"net/url"
	"os"
	"os/signal"

	"github.com/devusSs/minly/internal/clipboard"
	"github.com/devusSs/minly/internal/config"
	"github.com/devusSs/minly/internal/log"
	"github.com/devusSs/minly/internal/minio"
	"github.com/devusSs/minly/internal/secret"
	"github.com/devusSs/minly/internal/yourls"
	"github.com/spf13/cobra"
)

var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Uploads a file to MinIO and shortens the presigned URL",
	PreRun: func(_ *cobra.Command, _ []string) {
		if !uploadTest {
			err := log.Setup()
			checkErr(err, "failed to setup log package")

			go func() {
				err = log.CleanOld()
				if err != nil {
					log.Logger().Error().Err(err).Msg("failed to clean old log files")
				}
			}()
		}
	},
	PostRun: func(_ *cobra.Command, _ []string) {
		if !uploadTest {
			err := log.Flush()
			checkErr(err, "failed to flush log package")
		}
	},
	Run: func(_ *cobra.Command, args []string) {
		if uploadTest {
			log.Suppress()
			defer log.Enable()
		}

		if len(args) != 1 {
			logErr(errors.New("missing file path argument"), "file path argument is required")
		}

		filePath := args[0]
		if filePath == "" {
			logErr(errors.New("missing file path argument"), "file path argument is required")
		}

		log.Logger().Info().Str("file_path", filePath).
			Msg("file path argument received")

		var err error
		cfg, err = config.Read()
		logErr(err, "failed to read config file")

		log.Logger().Info().Any("config", cfg).
			Msg("config file read successfully")

		var minioAccessKey string
		minioAccessKey, err = getSecret(secret.MinioAccessKey)
		logErr(err, "failed to get MinIO access key secret")

		log.Logger().Info().Msg("got MinIO access key successfully")

		var minioAccessSecret string
		minioAccessSecret, err = getSecret(secret.MinioAccessSecret)
		logErr(err, "failed to get MinIO access secret")

		log.Logger().Info().Msg("got MinIO access secret successfully")

		var yourlsSignature string
		yourlsSignature, err = getSecret(secret.YOURLSignature)
		logErr(err, "failed to get YOURLS signature")

		log.Logger().Info().Msg("got YOURLS signature successfully")

		var mc *minio.Client
		mc, err = minio.NewClient(
			cfg.MinioEndpoint,
			minioAccessKey,
			minioAccessSecret,
			cfg.MinioUseSSL,
			cfg.MinioRegion,
		)
		logErr(err, "failed to create MinIO client")

		log.Logger().Info().
			Str("minio_endpoint", cfg.MinioEndpoint).
			Bool("minio_use_ssl", cfg.MinioUseSSL).
			Str("minio_region", cfg.MinioRegion).
			Msg("MinIO client created successfully")

		err = mc.Setup(cfg.MinioBucketName, cfg.MinioRegion, cfg.MinioLinkExpiry)
		logErr(err, "failed to setup MinIO client")

		log.Logger().Info().
			Str("minio_bucket_name", cfg.MinioBucketName).
			Str("minio_region", cfg.MinioRegion).
			Str("minio_link_expiry", cfg.MinioLinkExpiry.String()).
			Msg("MinIO client setup successfully")

		var yc *yourls.Client
		yc, err = yourls.NewClient(cfg.YOURLSEndpoint.String(), yourlsSignature)
		logErr(err, "failed to create YOURLS client")

		log.Logger().Info().Msg("YOURLS client created successfully")

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		defer cancel()

		var presignedURL *url.URL
		presignedURL, err = mc.UploadFile(ctx, filePath)
		logErr(err, "failed to upload file to MinIO")

		log.Logger().Info().Str("presigned_url", presignedURL.String()).
			Msg("file uploaded successfully")

		var shortURL string
		shortURL, err = yc.Shorten(ctx, presignedURL.String())
		logErr(err, "failed to shorten presigned URL using YOURLS")

		log.Logger().Info().Str("short_url", shortURL).
			Msg("presigned URL shortened successfully")

		if !uploadNoClip {
			err = clipboard.Write(shortURL)
			logErr(err, "failed to write short URL to clipboard")

			log.Logger().Info().Msg("short URL written to clipboard successfully")
		} else {
			log.Logger().Warn().Msg("short URL not written to clipboard due to --no-clip flag")
		}
	},
}

var (
	uploadTest   bool
	uploadNoClip bool
)

func init() {
	rootCmd.AddCommand(uploadCmd)

	uploadCmd.Flags().
		BoolVar(&uploadTest, "test", false, "run upload command in test mode (no logs)")
	uploadCmd.Flags().
		BoolVar(&uploadNoClip, "no-clip", false, "do not write short URL to clipboard")
}
