package lib

import (
	"github.com/albrow/5w4g-server/config"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
)

var s3Bucket *s3.Bucket

// S3Bucket returns an authenticated S3 Bucket instance
func S3Bucket() (*s3.Bucket, error) {

	if s3Bucket != nil {
		// If we've already got an authenticated bucket instance,
		// simply return it
		return s3Bucket, nil
	} else {
		// Otherwise we'll need to re-authenticate
		auth, err := aws.GetAuth(config.Aws.AccessKeyId, config.Aws.SecretAccessKey)
		if err != nil {
			return nil, err
		}
		client := s3.New(auth, aws.USEast)

		// Get the bucket by name
		s3Bucket = client.Bucket(config.Aws.BucketName)
		return s3Bucket, nil
	}
}
