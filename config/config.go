package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	BucketName     string
}

func LoadConfig() *Config {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Ошибка при загрузке .env файла")
	}
	return &Config{
		MinioEndpoint:  os.Getenv("MINIO_ENDPOINT"),
		MinioAccessKey: os.Getenv("MINIO_ACCESS_KEY"),
		MinioSecretKey: os.Getenv("MINIO_SECRET_KEY"),
		BucketName:     os.Getenv("MINIO_BUCKET"),
	}
}
