package entity

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_status_UnmarshalJSON(t *testing.T) {
	for i := 0; i < int(statusCompletedEnd); i++ {
		s := Status(i)
		marshal, err := json.Marshal(s)
		require.NoError(t, err)
		if string(marshal) == "\"unknown\"" {
			continue
		}
		var gotStatus Status
		err = json.Unmarshal(marshal, &gotStatus)
		require.NoError(t, err)
		require.Equal(t, s, gotStatus, string(marshal))
	}
}
