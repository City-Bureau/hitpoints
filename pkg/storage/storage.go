package storage

import (
	"time"
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
