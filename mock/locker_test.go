package mock

import (
	"context"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TestLocker_Lock(t *testing.T) {
	locker := NewMockLocker()
	lock, err := locker.Lock(context.Background(), "key", "value", time.Millisecond)
	assert.NilError(t, err)
	assert.Equal(t, lock, true)
	lock, err = locker.Lock(context.Background(), "key", "value", time.Millisecond)
	assert.NilError(t, err)
	assert.Equal(t, lock, false)
	time.Sleep(time.Millisecond)
	lock, err = locker.Lock(context.Background(), "key", "value", time.Millisecond)
	assert.NilError(t, err)
	assert.Equal(t, lock, true)
	unlock, err := locker.Unlock(context.Background(), "key")
	assert.NilError(t, err)
	assert.Equal(t, unlock, true)
	unlock, err = locker.Unlock(context.Background(), "key")
	assert.NilError(t, err)
	assert.Equal(t, unlock, false)
}
