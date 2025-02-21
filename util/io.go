package util

import "io"

func NopCloser(writer io.Writer) io.WriteCloser {
	return &nopCloser{writer}
}

type nopCloser struct {
	io.Writer
}

func (n *nopCloser) Close() error {
	return nil
}
