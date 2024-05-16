package util

// StringError is a simple error type that implements the error interface.
// It is useful when you want to return a constant error message.
type StringError string

func (e StringError) Error() string {
	return string(e)
}
