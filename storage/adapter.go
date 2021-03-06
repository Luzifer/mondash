package storage

import (
	"fmt"
	"net/url"

	"github.com/pkg/errors"
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
func GetStorage(uri string) (Storage, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, errors.Wrap(err, "Invalid storage URI")
	}

	switch u.Scheme {
	case "s3":
		return NewS3Storage(u), nil
	case "file":
		return NewFileStorage(u), nil
	}

	return nil, errors.Errorf("Storage %q not found", u.Scheme)
}
