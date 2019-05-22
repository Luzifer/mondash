package storage

import (
	"io/ioutil"
	"os"
	"path"

	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/mondash/config"
)

// FileStorage is a storage adapter storing the data into single local files
type FileStorage struct {
	cfg *config.Config
}

// NewFileStorage instanciates a new FileStorage
func NewFileStorage(cfg *config.Config) *FileStorage {
	// Create directory if not exists
	if _, err := os.Stat(cfg.FileStorage.Directory); os.IsNotExist(err) {
		if err := os.MkdirAll(cfg.FileStorage.Directory, 0700); err != nil {
			log.WithError(err).Fatal("Could not create storage directory")
		}
	}

	return &FileStorage{
		cfg: cfg,
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
	return path.Join(f.cfg.FileStorage.Directory, dashboardID+".txt")
}
