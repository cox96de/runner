package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/cockroachdb/errors"
	"github.com/cox96de/runner/util"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gotest.tools/v3/assert"
)

// NewMockDB creates a new database with tables created from given models.
// It uses sqlite in memory as the database.
func NewMockDB(t *testing.T, models ...interface{}) *Client {
	t.Helper()
	file := util.RandomID("sql-mocker")
	conn, err := gorm.Open(
		sqlite.Open(fmt.Sprintf("file:%s?mode=memory", file)),
		&gorm.Config{
			SkipDefaultTransaction: true,
			Logger:                 nil, // TODO: add custom logger.
		},
	)
	assert.NilError(t, err)
	err = migrateModels(conn, models...)
	assert.NilError(t, err)
	return NewClient(conn)
}

func TestClient_Transaction(t *testing.T) {
	client := NewMockDB(t, &Job{})
	t.Run("success", func(t *testing.T) {
		var createdJobs []*Job
		err := client.Transaction(func(client *Client) error {
			var err error
			createdJobs, err = client.CreateJobs(context.Background(), []*CreateJobOption{
				{
					PipelineID: 1,
				},
			})
			if err != nil {
				return err
			}
			return nil
		})
		assert.NilError(t, err)
		assert.Assert(t, len(createdJobs) > 0)
	})
	t.Run("failed", func(t *testing.T) {
		var createdJobs []*Job
		err := client.Transaction(func(client *Client) error {
			var err error
			createdJobs, err = client.CreateJobs(context.Background(), []*CreateJobOption{
				{
					PipelineID: 1,
				},
			})
			if err != nil {
				return err
			}
			return errors.New("some error")
		})
		assert.Error(t, err, "some error")
		_, err = client.GetJobByID(context.Background(), createdJobs[0].ID)
		assert.Assert(t, IsRecordNotFoundError(err))
	})
	t.Run("inner", func(t *testing.T) {
		var (
			createdJobs      []*Job
			innerCreatedJobs []*Job
		)
		err := client.Transaction(func(client *Client) error {
			var err error
			createdJobs, err = client.CreateJobs(context.Background(), []*CreateJobOption{
				{
					PipelineID: 1,
				},
			})
			if err != nil {
				return err
			}
			return client.Transaction(func(client *Client) error {
				innerCreatedJobs, err = client.CreateJobs(context.Background(), []*CreateJobOption{
					{
						PipelineID: 1,
					},
				})
				return err
			})
		})
		assert.NilError(t, err)
		assert.Assert(t, len(createdJobs) > 0)
		jobByID, err := client.GetJobByID(context.Background(), createdJobs[0].ID)
		assert.NilError(t, err)
		assert.DeepEqual(t, createdJobs[0], jobByID)
		assert.Assert(t, len(innerCreatedJobs) > 0)
		jobByID, err = client.GetJobByID(context.Background(), innerCreatedJobs[0].ID)
		assert.NilError(t, err)
		assert.DeepEqual(t, innerCreatedJobs[0], jobByID)
	})
}
