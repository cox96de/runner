package s3

//go:generate mockgen -destination mock/mockgen.go -package mock github.com/aws/aws-sdk-go/service/s3/s3iface S3API
