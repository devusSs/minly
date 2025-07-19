package minio

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

func (c *Client) UploadFile(ctx context.Context, path string) (*url.URL, error) {
	if !c.setup {
		return nil, errors.New("client is not set up")
	}

	if ctx == nil {
		return nil, errors.New("context cannot be nil")
	}

	if path == "" {
		return nil, errors.New("file path cannot be empty")
	}

	objectName, err := randomizeObjectName(path)
	if err != nil {
		return nil, fmt.Errorf("failed to randomize object name: %w", err)
	}

	var contentType string
	contentType, err = getContentType(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get content type: %w", err)
	}

	err = c.createBucketIfNotExists(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create bucket if not exists: %w", err)
	}

	_, err = c.minioClient.FPutObject(
		ctx,
		c.bucketName,
		objectName,
		path,
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	var presignedURL *url.URL
	presignedURL, err = c.generatePresignedURL(ctx, objectName)
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignedURL, nil
}

func randomizeObjectName(file string) (string, error) {
	if file == "" {
		return "", errors.New("file name cannot be empty")
	}

	file = filepath.Base(file)
	ext := filepath.Ext(file)

	uid, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("failed to generate random UUID: %w", err)
	}

	return fmt.Sprintf("%s%s", uid.String(), ext), nil
}

func getContentType(path string) (string, error) {
	if path == "" {
		return "", errors.New("file path cannot be empty")
	}

	mtype, err := mimetype.DetectFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to detect content type: %w", err)
	}

	return mtype.String(), nil
}
