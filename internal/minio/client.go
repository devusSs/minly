package minio

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Client struct {
	minioClient *minio.Client

	setup        bool
	bucketName   string
	bucketRegion string
	linkExpiry   time.Duration
}

func NewClient(
	endpoint string,
	accessKey string,
	accessSecret string,
	useSSL bool,
	region string,
) (*Client, error) {
	if endpoint == "" {
		return nil, errors.New("endpoint cannot be empty")
	}

	if accessKey == "" {
		return nil, errors.New("access key cannot be empty")
	}

	if accessSecret == "" {
		return nil, errors.New("access secret cannot be empty")
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, accessSecret, ""),
		Secure: useSSL,
		Region: region,
	})
	if err != nil {
		return nil, err
	}

	return &Client{minioClient: client, setup: false}, nil
}

func (c *Client) Setup(bucketName string, bucketRegion string, linkExpiry time.Duration) error {
	if c.setup {
		return errors.New("client is already set up")
	}

	if bucketName == "" {
		return errors.New("bucket name cannot be empty")
	}

	if bucketRegion == "" {
		return errors.New("bucket region cannot be empty")
	}

	c.bucketName = bucketName
	c.bucketRegion = bucketRegion
	c.linkExpiry = linkExpiry
	c.setup = true

	return nil
}

func (c *Client) createBucketIfNotExists(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context cannot be nil")
	}

	exists, err := c.minioClient.BucketExists(ctx, c.bucketName)
	if err != nil {
		return fmt.Errorf("failed to check if bucket exists: %w", err)
	}

	if !exists {
		err = c.minioClient.MakeBucket(ctx, c.bucketName, minio.MakeBucketOptions{
			Region: c.bucketRegion,
		})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return nil
}

func (c *Client) generatePresignedURL(ctx context.Context, objectName string) (*url.URL, error) {
	if ctx == nil {
		return nil, errors.New("context cannot be nil")
	}

	if objectName == "" {
		return nil, errors.New("object name cannot be empty")
	}

	presignedURL, err := c.minioClient.PresignedGetObject(
		ctx,
		c.bucketName,
		objectName,
		c.linkExpiry,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignedURL, nil
}
