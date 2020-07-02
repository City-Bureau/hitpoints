package storage

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

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
	err = os.MkdirAll(filepath.Join(s.path, hitsDir), 0760)
	if err != nil {
		log.Fatal(err)
	}
	return ioutil.WriteFile(filepath.Join(s.path, hitsDir, "hits.json"), recB, 0760)
}
