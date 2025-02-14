package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioStorage struct {
	client *minio.Client
	bucket string
}

func NewMinioStorage(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*MinioStorage, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к MinIO: %w", err)
	}

	return &MinioStorage{
		client: minioClient,
		bucket: bucket,
	}, nil
}

// GetFile возвращает поток данных из MinIO
func (s *MinioStorage) GetFile(objectName string) (io.ReadCloser, error) {
	obj, err := s.client.GetObject(context.Background(), s.bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("ошибка получения файла %s: %w", objectName, err)
	}
	return obj, nil
}
