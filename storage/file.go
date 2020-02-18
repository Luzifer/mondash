package storage

import (
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"sync"

	log "github.com/sirupsen/logrus"
)

// FileStorage is a storage adapter storing the data into single local files
type FileStorage struct {
	storagePath string

	dashLock     map[string]*sync.RWMutex
	dashLockLock *sync.Mutex
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

		dashLock:     map[string]*sync.RWMutex{},
		dashLockLock: new(sync.Mutex),
	}
}

// Put writes the given data to FS
func (f *FileStorage) Put(dashboardID string, data []byte) error {
	f.getLock(dashboardID).Lock()
	defer f.getLock(dashboardID).Unlock()

	err := ioutil.WriteFile(f.getFilePath(dashboardID), data, 0600)

	return err
}

// Get loads the data for the given dashboard from FS
func (f *FileStorage) Get(dashboardID string) ([]byte, error) {
	f.getLock(dashboardID).RLock()
	defer f.getLock(dashboardID).RUnlock()

	data, err := ioutil.ReadFile(f.getFilePath(dashboardID))
	if err != nil {
		return nil, DashboardNotFoundError{dashboardID}
	}

	return data, nil
}

// Delete deletes the given dashboard from FS
func (f *FileStorage) Delete(dashboardID string) error {
	f.getLock(dashboardID).Lock()
	defer f.getLock(dashboardID).Unlock()

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
	f.getLock(dashboardID).RLock()
	defer f.getLock(dashboardID).RUnlock()

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

func (f *FileStorage) getLock(dashboardID string) *sync.RWMutex {
	f.dashLockLock.Lock()
	defer f.dashLockLock.Unlock()

	l, ok := f.dashLock[dashboardID]
	if !ok {
		l = new(sync.RWMutex)
		f.dashLock[dashboardID] = l
	}

	return l
}
