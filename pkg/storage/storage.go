package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinioStorage структура для работы с MinIO
type MinioStorage struct {
	Client          *minio.Client
	Bucket          string
	Track_bucket    string
	Playlist_bucket string
}

// NewMinioStorage инициализация MinIO
func NewMinioStorage() (*MinioStorage, error) {
	// Загружаем переменные окружения из .env
	if err := godotenv.Load("pkg/storage/.env"); err != nil {
		log.Println("Предупреждение: .env файл не найден, используются системные переменные окружения")
	}

	// Читаем переменные окружения
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	bucket := os.Getenv("MINIO_BUCKET")
	track_bucket := os.Getenv("MINIO_TRACK_BUCKET")
	playlist_bucket := os.Getenv("MINIO_PLAYLIST_BUCKET")

	// Подключение к MinIO
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, err
	}

	log.Println("Подключение к MinIO успешно!")
	return &MinioStorage{
		Client:          client,
		Bucket:          bucket,
		Track_bucket:    track_bucket,
		Playlist_bucket: playlist_bucket,
	}, nil
}

// GetFile загружает объект из MinIO по имени файла
func (s *MinioStorage) GetFile(fileName string) (io.ReadCloser, error) {
	ctx := context.Background()
	obj, err := s.Client.GetObject(ctx, s.Bucket, fileName, minio.GetObjectOptions{})
	if err != nil {
		log.Printf("Ошибка при получении файла %s: %v", fileName, err)
		return nil, err
	}

	// // Пробуем прочитать первые байты, чтобы убедиться, что файл не пустой
	// buf := make([]byte, 512)
	// n, err := obj.Read(buf)
	// if err != nil && err != io.EOF {
	// 	log.Printf("Ошибка при чтении файла %s: %v", fileName, err)
	// 	return nil, err
	// }

	// if n == 0 {
	// 	log.Printf("Файл %s пуст", fileName)
	// 	return nil, errors.New("file is empty")
	// }

	// // Возвращаемся в начало файла
	// if _, err := obj.Seek(0, io.SeekStart); err != nil {
	// 	log.Printf("Ошибка при перемотке файла %s: %v", fileName, err)
	// 	return nil, err
	// }

	return obj, nil
}

func (s *MinioStorage) UploadFile(objectName, filePath string) error {
	// Открываем файл
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Загружаем файл в MinIO
	_, err = s.Client.PutObject(context.Background(), s.Bucket, objectName, file, -1, minio.PutObjectOptions{})
	return err
}

func (s *MinioStorage) UploadImage(bucketType, objectName string, file io.Reader) error {
	var err error
	if bucketType == "track" {
		_, err = s.Client.PutObject(context.Background(), s.Track_bucket, objectName, file, -1, minio.PutObjectOptions{})
	}
	if bucketType == "playlist" {
		_, err = s.Client.PutObject(context.Background(), s.Playlist_bucket, objectName, file, -1, minio.PutObjectOptions{})
	}
	return err
}

func (s *MinioStorage) GetImage(bucketType string, fileName string) (io.ReadCloser, error) {
	ctx := context.Background()
	var obj *minio.Object
	var err error

	// Проверяем, существует ли объект с помощью StatObject
	if bucketType == "track" {
		_, err = s.Client.StatObject(ctx, s.Track_bucket, fileName, minio.StatObjectOptions{})
	}
	if bucketType == "playlist" {
		_, err = s.Client.StatObject(ctx, s.Playlist_bucket, fileName, minio.StatObjectOptions{})
	}

	// Если объект не существует, вернем ошибку
	if err != nil {
		// Проверяем ошибку на отсутствие файла
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			log.Printf("Объект %s не найден в бакете %s", fileName, bucketType)
			return nil, fmt.Errorf("file not found")
		}
		log.Printf("Ошибка при получении метаданных для %s: %v", fileName, err)
		return nil, err
	}

	// Если объект существует, получаем его
	if bucketType == "track" {
		obj, err = s.Client.GetObject(ctx, s.Track_bucket, fileName, minio.GetObjectOptions{})
	}
	if bucketType == "playlist" {
		obj, err = s.Client.GetObject(ctx, s.Playlist_bucket, fileName, minio.GetObjectOptions{})
	}

	if err != nil {
		log.Printf("Ошибка при получении объекта %s: %v", fileName, err)
		return nil, err
	}

	log.Printf("Объект %s найден в бакете %s", fileName, bucketType)
	return obj, nil
}
