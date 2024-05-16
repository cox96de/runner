package agent

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func isErrorContextCancel(err error) bool {
	grpcError, ok := status.FromError(err)
	if ok {
		return grpcError.Code() == codes.DeadlineExceeded
	}
	return errors.Is(err, context.Canceled)
}
