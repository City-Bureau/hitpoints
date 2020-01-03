package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// StoreType is a wrapper type around a string to create an enum of storage options
type StoreType string

const (
	// File is a StoreType for saving to disk
	File StoreType = "file"
	// Azure is a StoreType for saving to Azure Blob Storage
	Azure StoreType = "azure"
	// S3 is a StoreType for saving to S3
	S3 StoreType = "s3"
)

// Record represents the structure for an individual archive file
type Record struct {
	Timestamp time.Time      `json:"timestamp"`
	Hits      map[string]int `json:"hits"`
}

// HitStorage is an interface for writing to a StoreType
type HitStorage interface {
	Archive(map[string]int) error
}

func datetimePath(t time.Time) string {
	return t.Format("2006/01/02/15/04/05")
}

// FileStorage manages writing archives to disk
type FileStorage struct {
	path string
}

// NewFileStorage creates a new FileStorage from a path
func NewFileStorage(path string) (HitStorage, error) {
	fileStorage := &FileStorage{path}
	return fileStorage, nil
}

// Archive saves an archive to local disk
func (s *FileStorage) Archive(items map[string]int) error {
	rec := &Record{
		Timestamp: time.Now().UTC(),
		Hits:      items,
	}

	recB, err := json.Marshal(rec)
	if err != nil {
		return err
	}

	hitsDir := datetimePath(rec.Timestamp)
	os.MkdirAll(hitsDir, 0644)
	return ioutil.WriteFile(filepath.Join(s.path, hitsDir, "hits.json"), recB, 0644)
}

// AzureStorage manages saving archives to Azure Blob Storage
type AzureStorage struct {
	accountName string
	accountKey  string
	container   string
	credential  *azblob.SharedKeyCredential
}

// NewAzureStorage creates a new AzureStorage from credentials
func NewAzureStorage(accountName string, accountKey string, container string) (HitStorage, error) {
	credential, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		return nil, err
	}
	azStorage := &AzureStorage{
		accountName,
		accountKey,
		container,
		credential,
	}
	return azStorage, nil
}

// Archive saves the cache to Azure Blob Storage
func (s *AzureStorage) Archive(items map[string]int) error {
	rec := &Record{
		Timestamp: time.Now().UTC(),
		Hits:      items,
	}

	recB, err := json.Marshal(rec)
	if err != nil {
		return err
	}

	blobURL, _ := url.Parse(fmt.Sprintf(
		"https://%s.blob.core.windows.net/%s/%s/hits.json",
		s.accountName,
		s.container,
		datetimePath(rec.Timestamp),
	))

	blockBlobURL := azblob.NewBlockBlobURL(
		*blobURL,
		azblob.NewPipeline(s.credential, azblob.PipelineOptions{}),
	)

	ctx := context.Background()
	_, err = azblob.UploadStreamToBlockBlob(
		ctx,
		bytes.NewReader(recB),
		blockBlobURL,
		azblob.UploadStreamToBlockBlobOptions{BufferSize: 2 * 1024 * 1024, MaxBuffers: 3},
	)

	return err
}

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
