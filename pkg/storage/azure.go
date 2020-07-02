package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/Azure/azure-storage-blob-go/azblob"
)

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
