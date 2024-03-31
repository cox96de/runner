package s3

import "github.com/aws/aws-sdk-go/service/s3/s3iface"

type Client struct {
	s3iface.S3API
}

func NewClient(s3 s3iface.S3API) *Client {
	return &Client{S3API: s3}
}
