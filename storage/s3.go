package storage // import "github.com/Luzifer/mondash/storage"

import (
	"io/ioutil"
	"strings"

	"github.com/Luzifer/mondash/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
)

type S3Storage struct {
	s3connection *s3.S3
	cfg          *config.Config
}

func NewS3Storage(cfg *config.Config) *S3Storage {
	s3connection := s3.New(&aws.Config{})
	return &S3Storage{
		s3connection: s3connection,
	}
}

func (s *S3Storage) Put(dashboardID string, data []byte) error {
	_, err := s.s3connection.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(s.cfg.S3.Bucket),
		ContentType: aws.String("application/json"),
		Key:         aws.String(dashboardID),
		// TODO: Private ACL
	})

	return err
}

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

func (s *S3Storage) Delete(dashboardID string) error {
	_, err := s.s3connection.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.cfg.S3.Bucket),
		Key:    aws.String(dashboardID),
	})

	return err
}

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
		} else {
			return false, err
		}
	}

	return true, nil
}
