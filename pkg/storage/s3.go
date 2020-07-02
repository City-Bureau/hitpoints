package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// S3Storage manages saving files to S3
type S3Storage struct {
	accessKeyID     string
	secretAccessKey string
	bucket          string
	svc             *s3.S3
}

// NewS3Storage creates a new S3Storage instance
func NewS3Storage(accessKeyID string, secretAccessKey string, bucket string, region string, useRole bool) (HitStorage, error) {
	var creds *credentials.Credentials
	if useRole {
		sess, _ := session.NewSession(&aws.Config{
			Region: aws.String(region),
		})
		creds = ec2rolecreds.NewCredentials(sess)
	} else if accessKeyID != "" && secretAccessKey != "" {
		creds = credentials.NewStaticCredentials(accessKeyID, secretAccessKey, "")
	} else {
		creds = credentials.NewEnvCredentials()
	}
	svc := s3.New(session.Must(session.NewSession(&aws.Config{
		Credentials: creds,
		Region:      aws.String(region),
	})))
	s3Storage := &S3Storage{
		accessKeyID,
		secretAccessKey,
		bucket,
		svc,
	}
	return s3Storage, nil
}

// Archive saves file to S3
func (s *S3Storage) Archive(items map[string]int) error {
	rec := &Record{
		Timestamp: time.Now().UTC(),
		Hits:      items,
	}

	recB, err := json.Marshal(rec)
	if err != nil {
		return err
	}

	s3Input := &s3.PutObjectInput{
		Body:   aws.ReadSeekCloser(bytes.NewReader(recB)),
		Bucket: aws.String(s.bucket),
		Key:    aws.String(fmt.Sprintf("%s/hits.json", datetimePath(rec.Timestamp))),
	}

	_, err = s.svc.PutObject(s3Input)

	return err
}
