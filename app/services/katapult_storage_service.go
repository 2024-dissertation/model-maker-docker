package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type KatapultStorageService struct {
	client     *s3.Client
	bucketName string
	region     string
	endpoint   string
}

func NewKatapultStorageService() *KatapultStorageService {
	bucketName := os.Getenv("KATAPULT_BUCKET_NAME")
	region := os.Getenv("KATAPULT_REGION")
	endpoint := os.Getenv("KATAPULT_ENDPOINT")
	accessKey := os.Getenv("KATAPULT_ACCESS_KEY")
	secretKey := os.Getenv("KATAPULT_SECRET_KEY")

	// Configure AWS SDK
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		panic(fmt.Sprintf("unable to load SDK config, %v", err))
	}

	// Create S3 client with custom endpoint
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})

	return &KatapultStorageService{
		client:     client,
		bucketName: bucketName,
		region:     region,
		endpoint:   endpoint,
	}
}

func (s *KatapultStorageService) UploadFile(file *multipart.FileHeader, taskID uint, fileType string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	return s.UploadFromReader(src, taskID, file.Filename, fileType)
}

func (s *KatapultStorageService) UploadFromReader(reader io.Reader, taskID uint, filename string, fileType string) (string, error) {
	// Read the entire file into memory
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, reader); err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	objectKey := s.getObjectKey(taskID, filename, fileType)

	// Upload to S3
	_, err := s.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(objectKey),
		Body:   bytes.NewReader(buf.Bytes()),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	return objectKey, nil
}

func (s *KatapultStorageService) GetFile(filepath string) (io.ReadCloser, error) {
	// Get object from S3
	result, err := s.client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(filepath),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	return result.Body, nil
}

func (s *KatapultStorageService) DeleteFile(taskID uint, filename string) error {
	objectKey := s.getObjectKey(taskID, filename, "")

	// Delete object from S3
	_, err := s.client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

func (s *KatapultStorageService) getObjectKey(taskID uint, filename string, fileType string) string {
	if fileType == "mesh" {
		return fmt.Sprintf("objects/%d/%s", taskID, filename)
	}
	return fmt.Sprintf("uploads/%d/%s", taskID, filename)
}
