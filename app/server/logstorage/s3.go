package logstorage

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

type S3 struct {
	client s3iface.S3API
	bucket string
}

func NewS3(bucket string, client s3iface.S3API) *S3 {
	return &S3{bucket: bucket, client: client}
}

func (s *S3) Open(ctx context.Context, filename string) (io.ReadCloser, error) {
	getObjectOutput, err := s.client.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: &s.bucket,
		Key:    &filename,
	})
	if err != nil {
		return nil, err
	}
	return getObjectOutput.Body, nil
}

func (s *S3) Save(ctx context.Context, filename string, r Reader) error {
	_, err := s.client.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket: &s.bucket,
		Key:    &filename,
		Body:   r,
	})
	if err != nil {
		return err
	}
	return nil
}
