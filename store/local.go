package store

import (
	"os"
	"path/filepath"
)

// Local is local store
type Local struct {
	f *os.File
}

// NewLocal creates local store
func NewLocal(path string) (*Local, error) {
	// Crete directory structure
	if err := os.MkdirAll(filepath.Base(path), 0755); err != nil {
		return nil, err
	}
	// create tile in path
	filePath := filepath.Clean(path)
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}

	return &Local{f}, nil
}

// Write writes data to local store
func (l *Local) Write(b []byte) (n int, err error) {
	return l.f.Write(b)
}

// Read reads data from local store
func (l *Local) Read(b []byte) (n int, err error) {
	return l.f.Read(b)
}
