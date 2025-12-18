package storage

import (
	"context"
	"fmt"
	"io"
	"path"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOStorage implements Storage using MinIO as the backend.
type MinIOStorage struct {
	client    *minio.Client
	bucket    string
	namespace string
}

// MinIOConfig holds the configuration for MinIO connection.
type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

// NewMinIOStorage creates a new MinIO storage instance.
func NewMinIOStorage(cfg MinIOConfig) (*MinIOStorage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	return &MinIOStorage{
		client: client,
		bucket: cfg.Bucket,
	}, nil
}

// EnsureBucket creates the bucket if it doesn't exist.
func (s *MinIOStorage) EnsureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return nil
}

// objectKey returns the full object key including namespace prefix.
func (s *MinIOStorage) objectKey(key string) string {
	if s.namespace == "" {
		return key
	}
	return path.Join(s.namespace, key)
}

// Get retrieves a cache entry from MinIO.
func (s *MinIOStorage) Get(ctx context.Context, key string) (io.ReadCloser, int64, error) {
	objectKey := s.objectKey(key)

	obj, err := s.client.GetObject(ctx, s.bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get object: %w", err)
	}

	stat, err := obj.Stat()
	if err != nil {
		obj.Close()
		errResp := minio.ToErrorResponse(err)
		if errResp.Code == "NoSuchKey" {
			return nil, 0, ErrNotFound
		}
		return nil, 0, fmt.Errorf("failed to stat object: %w", err)
	}

	return obj, stat.Size, nil
}

// Put stores a cache entry in MinIO.
func (s *MinIOStorage) Put(ctx context.Context, key string, reader io.Reader, size int64) error {
	objectKey := s.objectKey(key)

	_, err := s.client.PutObject(ctx, s.bucket, objectKey, reader, size, minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		return fmt.Errorf("failed to put object: %w", err)
	}

	return nil
}

// Exists checks if a cache entry exists in MinIO.
func (s *MinIOStorage) Exists(ctx context.Context, key string) (bool, error) {
	objectKey := s.objectKey(key)

	_, err := s.client.StatObject(ctx, s.bucket, objectKey, minio.StatObjectOptions{})
	if err != nil {
		errResp := minio.ToErrorResponse(err)
		if errResp.Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("failed to stat object: %w", err)
	}

	return true, nil
}

// Delete removes a cache entry from MinIO.
func (s *MinIOStorage) Delete(ctx context.Context, key string) error {
	objectKey := s.objectKey(key)

	err := s.client.RemoveObject(ctx, s.bucket, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to remove object: %w", err)
	}

	return nil
}

// Ping checks if MinIO is reachable.
func (s *MinIOStorage) Ping(ctx context.Context) error {
	_, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("failed to ping MinIO: %w", err)
	}
	return nil
}

// WithNamespace returns a new MinIOStorage instance scoped to the given namespace.
func (s *MinIOStorage) WithNamespace(namespace string) Storage {
	return &MinIOStorage{
		client:    s.client,
		bucket:    s.bucket,
		namespace: namespace,
	}
}
