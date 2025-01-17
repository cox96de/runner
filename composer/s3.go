package composer

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/samber/lo"
)

type S3 struct {
	Endpoint         string `mapstructure:"endpoint" yaml:"endpoint"`
	Region           string `mapstructure:"region" yaml:"region"`
	S3ForcePathStyle bool   `mapstructure:"s3_force_path_style" yaml:"s3_force_path_style"`
	AccessKeyID      string `mapstructure:"access_key_id" yaml:"access_key_id"`
	SecretAccessKey  string `mapstructure:"secret_access_key" yaml:"secret_access_key"`
}

func ComposeS3(c *S3) (*s3.S3, error) {
	config := aws.Config{
		Credentials: credentials.NewCredentials(&credentials.StaticProvider{
			Value: credentials.Value{
				AccessKeyID:     c.AccessKeyID,
				SecretAccessKey: c.SecretAccessKey,
			},
		}),
		S3ForcePathStyle: lo.ToPtr(c.S3ForcePathStyle),
		Region:           &c.Region,
		Endpoint:         &c.Endpoint,
	}
	sessionWithOptions, err := session.NewSessionWithOptions(session.Options{Config: config})
	if err != nil {
		return nil, err
	}
	return s3.New(sessionWithOptions), err
}
