package util

import "io"

type nopCloser struct {
	io.Writer
}

func NopCloser(w io.Writer) io.WriteCloser {
	return &nopCloser{Writer: w}
}

func (n *nopCloser) Close() error {
	return nil
}
