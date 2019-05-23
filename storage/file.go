package storage

import (
	"io/ioutil"
	"net/url"
	"os"
	"path"

	log "github.com/sirupsen/logrus"
)

// FileStorage is a storage adapter storing the data into single local files
type FileStorage struct {
	storagePath string
}

// NewFileStorage instanciates a new FileStorage
func NewFileStorage(uri *url.URL) *FileStorage {
	// Create directory if not exists
	if _, err := os.Stat(uri.Path); os.IsNotExist(err) {
		if err := os.MkdirAll(uri.Path, 0700); err != nil {
			log.WithError(err).Fatal("Could not create storage directory")
		}
	}

	return &FileStorage{
		storagePath: uri.Path,
	}
}

// Put writes the given data to FS
func (f *FileStorage) Put(dashboardID string, data []byte) error {
	err := ioutil.WriteFile(f.getFilePath(dashboardID), data, 0600)

	return err
}

// Get loads the data for the given dashboard from FS
func (f *FileStorage) Get(dashboardID string) ([]byte, error) {
	data, err := ioutil.ReadFile(f.getFilePath(dashboardID))
	if err != nil {
		return nil, DashboardNotFoundError{dashboardID}
	}

	return data, nil
}

// Delete deletes the given dashboard from FS
func (f *FileStorage) Delete(dashboardID string) error {
	if exists, err := f.Exists(dashboardID); err != nil || !exists {
		if err != nil {
			return err
		}
		return DashboardNotFoundError{dashboardID}
	}

	return os.Remove(f.getFilePath(dashboardID))
}

// Exists checks for the existence of the given dashboard
func (f *FileStorage) Exists(dashboardID string) (bool, error) {
	if _, err := os.Stat(f.getFilePath(dashboardID)); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (f *FileStorage) getFilePath(dashboardID string) string {
	return path.Join(f.storagePath, dashboardID+".txt")
}
