package storage

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/Luzifer/mondash/config"
)

// Storage is an interface to have all storage systems compatible to each other
type Storage interface {
	Put(dashboardID string, data []byte) error
	Get(dashboardID string) ([]byte, error)
	Delete(dashboardID string) error
	Exists(dashboardID string) (bool, error)
}

// DashboardNotFoundError signalizes the requested dashboard could not be found
type DashboardNotFoundError struct {
	DashboardID string
}

func (e DashboardNotFoundError) Error() string {
	return fmt.Sprintf("Dashboard with ID '%s' was not found.", e.DashboardID)
}

// GetStorage acts as a storage factory providing the storage named by input
// name parameter
func GetStorage(cfg *config.Config) (Storage, error) {
	switch cfg.Storage {
	case "s3":
		return NewS3Storage(cfg), nil
	case "file":
		return NewFileStorage(cfg), nil
	}

	return nil, errors.Errorf("Storage %q not found", cfg.Storage)
}
