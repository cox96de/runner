package lib

import (
	"testing"

	"github.com/google/go-cmp/cmp/cmpopts"
	jsoniter "github.com/json-iterator/go"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gotest.tools/v3/assert"
)

func TestProtobufTypeExtension(t *testing.T) {
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	json.RegisterExtension(&ProtobufTypeExtension{})
	bptime := timestamppb.Now()
	m := map[string]*timestamppb.Timestamp{
		"now": bptime,
	}
	marshal, err := json.Marshal(m)
	assert.NilError(t, err)
	bb := map[string]*timestamppb.Timestamp{}
	err = json.Unmarshal(marshal, &bb)
	assert.NilError(t, err)
	assert.DeepEqual(t, m, bb, cmpopts.IgnoreUnexported(timestamppb.Timestamp{}))
}
