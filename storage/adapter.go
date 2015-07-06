package storage // import "github.com/Luzifer/mondash/storage"

import (
	"fmt"

	"github.com/Luzifer/mondash/config"
)

// Storage is an interface to have all storage systems compatible to each other
type Storage interface {
	Put(dashboardID string, data []byte) error
	Get(dashboardID string) ([]byte, error)
	Delete(dashboardID string) error
	Exists(dashboardID string) (bool, error)
}

// NotFoundError is a named error for more simple determination which
// type of error is thrown
type NotFoundError struct {
	Name string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("Storage '%s' not found.", e.Name)
}

// GetStorage acts as a storage factory providing the storage named by input
// name parameter
func GetStorage(cfg *config.Config) (Storage, error) {
	switch cfg.Storage {
	case "s3":
		return NewS3Storage(cfg), nil
	}

	return nil, NotFoundError{cfg.Storage}
}
