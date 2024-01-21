package executorpb

import (
	"bytes"
	"io"
)

func ReadAllFromCommandLog(s Executor_GetCommandLogClient) (string, error) {
	r := &bytes.Buffer{}
	for {
		recv, err := s.Recv()
		if err != nil {
			if err == io.EOF {
				return r.String(), nil
			}
			return r.String(), err
		}
		r.Write(recv.Output)
	}
}
