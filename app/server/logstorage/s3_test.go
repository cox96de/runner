package logstorage

import (
	"bytes"
	"context"
	"github.com/cox96de/runner/composer"
	"gotest.tools/v3/assert"
	"testing"
)

func ComposeS3Client() *S3 {
	composeS3, err := composer.ComposeS3(&composer.S3{
		Endpoint:         "http://127.0.0.1:9002",
		Region:           "CN",
		S3ForcePathStyle: true,
		AccessKeyID:      "xia4e1WQOGYHPjrCkiAj",
		SecretAccessKey:  "KUuP6RTHxN3980WPeKtw4nFO2KgoJorTtVc69ORE",
	})
	if err != nil {
		panic(err)
	}
	return NewS3("runner", composeS3)

}

func TestNewS3(t *testing.T) {
	client := ComposeS3Client()
	err := client.Save(context.Background(), "test", bytes.NewReader([]byte("content")))
	assert.NilError(t, err)
	open, err := client.Open(context.Background(), "test")
	assert.NilError(t, err)
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(open)
	assert.NilError(t, err)
	assert.Equal(t, buf.String(), "content")
}
