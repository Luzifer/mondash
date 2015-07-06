package storage // import "github.com/Luzifer/mondash/storage"

import (
	"bytes"
	"io/ioutil"
	"strings"

	"github.com/Luzifer/mondash/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
)

// S3Storage is a storage adapter storing the data into single S3 files
type S3Storage struct {
	s3connection *s3.S3
	cfg          *config.Config
}

// NewS3Storage instanciates a new S3Storage
func NewS3Storage(cfg *config.Config) *S3Storage {
	s3connection := s3.New(&aws.Config{})
	return &S3Storage{
		s3connection: s3connection,
		cfg:          cfg,
	}
}

// Put writes the given data to S3
func (s *S3Storage) Put(dashboardID string, data []byte) error {
	_, err := s.s3connection.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(s.cfg.S3.Bucket),
		ContentType: aws.String("application/json"),
		Key:         aws.String(dashboardID),
		Body:        bytes.NewReader(data),
		ACL:         aws.String("private"),
	})

	return err
}

// Get loads the data for the given dashboard from S3
func (s *S3Storage) Get(dashboardID string) ([]byte, error) {
	res, err := s.s3connection.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.cfg.S3.Bucket),
		Key:    aws.String(dashboardID),
	})
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// Delete deletes the given dashboard from S3
func (s *S3Storage) Delete(dashboardID string) error {
	_, err := s.s3connection.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.cfg.S3.Bucket),
		Key:    aws.String(dashboardID),
	})

	return err
}

// Exists checks for the existence of the given dashboard
func (s *S3Storage) Exists(dashboardID string) (bool, error) {
	_, err := s.s3connection.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(s.cfg.S3.Bucket),
		Key:    aws.String(dashboardID),
	})

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if strings.Contains(awsErr.Error(), "status code: 404") {
				return false, nil
			}
			return false, err
		}
		return false, err
	}

	return true, nil
}
