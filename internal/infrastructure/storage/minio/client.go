package minio

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Client struct {
	client     *minio.Client
	bucketName string
}

func NewClient(endpoint, accessKey, secretKey, bucketName string, useSSL bool) (*Client, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket: %w", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
		slog.Info("Created bucket", "bucket", bucketName)
	}

	return &Client{
		client:     client,
		bucketName: bucketName,
	}, nil
}

func (c *Client) UploadFile(ctx context.Context, fragmentID, fileType string, filename string, reader io.Reader, size int64) (string, error) {
	objectName := filepath.Join(fragmentID, fileType, filename)
	_, err := c.client.PutObject(ctx, c.bucketName, objectName, reader, size, minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	return objectName, nil
}

func (c *Client) GetFile(ctx context.Context, objectName string) (io.ReadCloser, int64, string, error) {
	obj, err := c.client.GetObject(ctx, c.bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, 0, "", fmt.Errorf("failed to get object: %w", err)
	}

	stat, err := obj.Stat()
	if err != nil {
		obj.Close()
		return nil, 0, "", fmt.Errorf("failed to stat object: %w", err)
	}

	return obj, stat.Size, stat.ContentType, nil
}

func (c *Client) GetPresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	url, err := c.client.PresignedGetObject(ctx, c.bucketName, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get presigned url: %w", err)
	}

	return url.String(), nil
}

func (c *Client) UploadAvatar(ctx context.Context, userID, filename string, reader io.Reader, size int64) (string, error) {
	objectName := fmt.Sprintf("avatars/%s/%s", userID, filename)
	_, err := c.client.PutObject(ctx, c.bucketName, objectName, reader, size, minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload avatar: %w", err)
	}

	return objectName, nil
}

func (c *Client) DeleteFiles(ctx context.Context, fragmentID string) error {
	objectsCh := make(chan minio.ObjectInfo)

	go func() {
		defer close(objectsCh)
		for object := range c.client.ListObjects(ctx, c.bucketName, minio.ListObjectsOptions{
			Prefix:    fragmentID + "/",
			Recursive: true,
		}) {
			if object.Err != nil {
				slog.Error("Error listing object", "error", object.Err)
				continue
			}
			objectsCh <- object
		}
	}()

	errorCh := c.client.RemoveObjects(ctx, c.bucketName, objectsCh, minio.RemoveObjectsOptions{})

	for err := range errorCh {
		if err.Err != nil {
			slog.Error("Error removing object", "error", err.Err)
		}
	}

	return nil
}