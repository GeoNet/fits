// Package s3 is for working with AWS S3 buckets.
package s3

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
	"os"
)

type S3 struct {
	client *s3.S3
}

// New returns an S3 using the default AWS credentials chain.
// This consults (in order) environment vars, config files, ec2 and ecs roles.
// It is an error if the AWS_REGION environment variable is not set.
// Requests with recoverable errors will be retried with the default retrier
func New() (S3, error) {
	if os.Getenv("AWS_REGION") == "" {
		return S3{}, errors.New("AWS_REGION is not set")
	}

	s, err := session.NewSession()
	if err != nil {
		return S3{}, errors.WithStack(err)
	}

	return S3{client: s3.New(s)}, nil
}

func NewWithCreds(creds *credentials.Credentials) (S3, error) {
	if os.Getenv("AWS_REGION") == "" {
		return S3{}, errors.New("AWS_REGION is not set")
	}

	s, err := session.NewSession()
	if err != nil {
		return S3{}, errors.WithStack(err)
	}

	return S3{client: s3.New(s, &aws.Config{Credentials: creds})}, nil
}

func NewAnonymous(region string) (S3, error) {
	if region == "" {
		region = os.Getenv("AWS_REGION")
	}
	if region == "" {
		return S3{}, errors.New("AWS_REGION is not set")
	}

	s, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.AnonymousCredentials,
	})
	if err != nil {
		return S3{}, errors.WithStack(err)
	}

	return S3{client: s3.New(s)}, nil
}

// Get gets the object referred to by key and version from bucket and write is into b.
// version can be zero.
func (s *S3) Get(bucket, key, version string, b *bytes.Buffer) error {
	params := s3.GetObjectInput{
		Key:    aws.String(key),
		Bucket: aws.String(bucket),
	}

	if version != "" {
		params.VersionId = aws.String(version)
	}

	result, err := s.client.GetObject(&params)
	if err != nil {
		return errors.WithStack(err)
	}
	defer result.Body.Close()

	_, err = b.ReadFrom(result.Body)

	return err
}

// Put puts the object to key in bucket.
func (s *S3) Put(bucket, key string, object []byte) error {
	input := s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(object),
	}

	_, err := s.client.PutObject(&input)
	return err
}

// Exists checks if an object for key already exists in the bucket.
func (s *S3) Exists(bucket, key string) (bool, error) {
	_, err := s.client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err == nil {
		return true, nil
	}

	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case "NotFound":
			return false, nil
		default:
			return false, errors.New("Error is not 'NotFound': " + aerr.Error())
		}
	}

	return false, errors.New("Unable to determine error: " + err.Error())
}

//Returns a list of object keys that match the provided prefix
func (s *S3) List(bucket, prefix string) ([]string, error) {
	result := make([]string, 0)

	input := &s3.ListObjectsInput{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	}

	out, err := s.client.ListObjects(input)
	if err != nil {
		return nil, err
	}

	for _, o := range out.Contents {
		result = append(result, *o.Key)
	}

	return result, nil
}

//Returns a list of object that match the provided prefix
func (s *S3) ListObjects(bucket, prefix string) ([]*s3.Object, error) {
	input := &s3.ListObjectsInput{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	}

	out, err := s.client.ListObjects(input)
	if err != nil {
		return nil, err
	}

	return out.Contents, nil
}

func (s *S3) Delete(bucket, key string) error {
	input := s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	_, err := s.client.DeleteObject(&input)
	return err
}
