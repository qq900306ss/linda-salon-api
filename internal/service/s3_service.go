package service

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type S3Service struct {
	client     *s3.Client
	bucketName string
	region     string
}

func NewS3Service() (*S3Service, error) {
	bucketName := os.Getenv("AWS_S3_BUCKET")
	if bucketName == "" {
		bucketName = "linda-salon-assets"
	}

	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "ap-northeast-1"
	}

	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	var cfg aws.Config
	var err error

	if accessKey != "" && secretKey != "" {
		// 使用明確的憑證
		cfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		)
	} else {
		// 使用預設憑證鏈 (IAM role, environment, etc.)
		cfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(region),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	return &S3Service{
		client:     client,
		bucketName: bucketName,
		region:     region,
	}, nil
}

// UploadFile 上傳檔案到 S3
func (s *S3Service) UploadFile(ctx context.Context, file *multipart.FileHeader, folder string) (string, error) {
	// 打開檔案
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	// 產生唯一檔名
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s-%s%s", uuid.New().String(), time.Now().Format("20060102150405"), ext)
	key := filepath.Join(folder, filename)
	key = strings.ReplaceAll(key, "\\", "/") // 確保使用正斜線

	// 讀取檔案內容
	fileBytes, err := io.ReadAll(src)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// 上傳到 S3
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		Body:        strings.NewReader(string(fileBytes)),
		ContentType: aws.String(file.Header.Get("Content-Type")),
		ACL:         "public-read", // 公開讀取
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	// 返回 S3 URL
	url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucketName, s.region, key)
	return url, nil
}

// DeleteFile 從 S3 刪除檔案
func (s *S3Service) DeleteFile(ctx context.Context, fileURL string) error {
	// 從 URL 提取 key
	key := s.extractKeyFromURL(fileURL)
	if key == "" {
		return fmt.Errorf("invalid S3 URL")
	}

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		return fmt.Errorf("failed to delete from S3: %w", err)
	}

	return nil
}

// extractKeyFromURL 從 S3 URL 提取 key
func (s *S3Service) extractKeyFromURL(url string) string {
	// 支援格式:
	// https://bucket.s3.region.amazonaws.com/path/to/file
	// https://bucket.s3.amazonaws.com/path/to/file
	prefix := fmt.Sprintf("https://%s.s3", s.bucketName)
	if !strings.HasPrefix(url, prefix) {
		return ""
	}

	parts := strings.SplitN(url, ".amazonaws.com/", 2)
	if len(parts) != 2 {
		return ""
	}

	return parts[1]
}
