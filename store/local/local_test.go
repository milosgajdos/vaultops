package local

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStore(t *testing.T) {
	dir := os.TempDir()
	fileName := "file.tmp"
	path := filepath.Join(dir, fileName)

	s, err := NewStore(path)
	defer os.Remove(path)
	assert.NoError(t, err)
	assert.Equal(t, path, s.f.Name())

	dir = "/etc"
	fileName = "passwd"
	s, err = NewStore(filepath.Join(dir, fileName))
	assert.Error(t, err)
}

func TestWriteRead(t *testing.T) {
	dir := os.TempDir()
	fileName := "file2.tmp"
	path := filepath.Join(dir, fileName)

	s, err := NewStore(path)
	defer os.Remove(path)
	assert.NoError(t, err)
	assert.Equal(t, path, s.f.Name())

	data := []byte("testdata")
	n, err := s.Write(data)
	assert.NoError(t, err)
	assert.Equal(t, len(data), n)
	// close the file so the write buffer is flushed
	assert.NoError(t, s.f.Close())

	s, err = NewStore(path)
	bufRead := make([]byte, len(data))
	n, err = s.Read(bufRead)
	assert.NoError(t, err)
	assert.Equal(t, n, len(data))
}
