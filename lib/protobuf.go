package lib

import (
	"encoding/json"
	"time"
	"unsafe"

	jsoniter "github.com/json-iterator/go"
	"github.com/modern-go/reflect2"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var timestampType = reflect2.TypeOf((*timestamppb.Timestamp)(nil))

type ProtobufTypeExtension struct {
	jsoniter.DummyExtension
}

func (e *ProtobufTypeExtension) CreateEncoder(typ reflect2.Type) jsoniter.ValEncoder {
	if typ == timestampType {
		return &timestampCodec{}
	}
	return nil
}

func (e *ProtobufTypeExtension) CreateDecoder(typ reflect2.Type) jsoniter.ValDecoder {
	if typ == timestampType {
		return &timestampCodec{}
	}
	return nil
}

type timestampCodec struct{}

func (c *timestampCodec) Decode(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
	var t time.Time
	iter.ReadVal(&t)
	tt := (**timestamppb.Timestamp)(ptr)
	ttt := timestamppb.New(t)
	*tt = ttt
}

func (c *timestampCodec) IsEmpty(ptr unsafe.Pointer) bool {
	return (*timestamppb.Timestamp)(ptr) == nil
}

func (c *timestampCodec) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	t := (**timestamppb.Timestamp)(ptr)
	marshal, err := json.Marshal((*t).AsTime())
	if err != nil {
		stream.Error = err
		return
	}
	_, err = stream.Write(marshal)
	if err != nil {
		stream.Error = err
	}
}
