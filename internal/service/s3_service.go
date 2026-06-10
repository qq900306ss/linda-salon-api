package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

// ErrUploadNotConfigured is returned when S3_BUCKET is not set.
var ErrUploadNotConfigured = errors.New("upload is not configured (S3_BUCKET missing)")

// PresignExpiry is how long a presigned upload URL stays valid.
const PresignExpiry = 5 * time.Minute

// S3Service issues presigned PUT URLs for image uploads.
type S3Service struct {
	bucket  string
	region  string
	presign *s3.PresignClient
}

// NewS3Service creates an S3Service. bucket may be empty, in which case
// PresignUpload returns ErrUploadNotConfigured.
func NewS3Service(ctx context.Context, bucket string) (*S3Service, error) {
	if bucket == "" {
		return &S3Service{}, nil
	}
	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}
	client := s3.NewFromConfig(cfg)
	return &S3Service{
		bucket:  bucket,
		region:  cfg.Region,
		presign: s3.NewPresignClient(client),
	}, nil
}

// PresignUpload returns a presigned PUT URL and the resulting public URL.
func (s *S3Service) PresignUpload(ctx context.Context, fileName, contentType string) (uploadURL, publicURL string, err error) {
	if s.bucket == "" {
		return "", "", ErrUploadNotConfigured
	}
	key := fmt.Sprintf("uploads/%s-%s", uuid.NewString(), fileName)
	req, err := s.presign.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(PresignExpiry))
	if err != nil {
		return "", "", fmt.Errorf("presign put object: %w", err)
	}
	publicURL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucket, s.region, key)
	return req.URL, publicURL, nil
}
