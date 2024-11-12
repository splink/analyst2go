package util

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"log"
)

func UploadToS3(key string, data []byte) (string, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(Env("AWS_REGION"))},
	)
	if err != nil {
		log.Println("Error creating s3 session", err)
		return "", err
	}

	s3Client := s3.New(sess)

	bucketName := Env("AWS_BUCKET")
	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	filePath := fmt.Sprintf("s3://%s/%s", bucketName, key)
	return filePath, nil
}
