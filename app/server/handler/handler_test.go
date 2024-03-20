package handler

import (
	"testing"
)

func TestHandler_mustEmbedUnimplementedServerServer(t *testing.T) {
	// Just increase test coverage.
	h := &Handler{}
	h.mustEmbedUnimplementedServerServer()
}
