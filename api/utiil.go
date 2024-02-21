package api

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type Time interface {
	time.Time | *time.Time
}

func ConvertTime[T Time](t T) *timestamppb.Timestamp {
	tt := interface{}(t)
	switch tb := tt.(type) {
	case time.Time:
		return timestamppb.New(tb)
	case *time.Time:
		if tb == nil {
			return nil
		}
		return timestamppb.New(*tb)
	}
	return nil
}
