package main

func increaseMaxOpenFiles() error {
	// There is no need to increase the max open files on Windows.
	return nil
}
