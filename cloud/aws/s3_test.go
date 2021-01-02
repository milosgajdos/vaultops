package aws

import (
	"fmt"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/stretchr/testify/assert"
)

func TestNewS3(t *testing.T) {
	s, err := NewS3("bucket", "key")
	assert.NotNil(t, s)
	assert.NoError(t, err)
}

type mockS3 struct {
	UploadFunc   func(*s3manager.UploadInput, ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)
	DownloadFunc func(io.WriterAt, *s3.GetObjectInput, ...func(*s3manager.Downloader)) (int64, error)
}

func (m *mockS3) Upload(in *s3manager.UploadInput, opts ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
	return m.UploadFunc(in, opts...)
}

func (m *mockS3) Download(w io.WriterAt, in *s3.GetObjectInput, opts ...func(*s3manager.Downloader)) (int64, error) {
	return m.DownloadFunc(w, in, opts...)
}

func TestWrite(t *testing.T) {
	c := &mockS3{}
	s3Client := S3{uploader: c, bucket: "bucket", key: "key"}

	data := []byte("testdata")
	// success uploading
	c.UploadFunc = func(in *s3manager.UploadInput, opts ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
		return &s3manager.UploadOutput{Location: "awsLocation"}, nil
	}

	_, err := s3Client.Write(data)
	assert.NoError(t, err)

	// error uploading
	c.UploadFunc = func(in *s3manager.UploadInput, opts ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
		return nil, fmt.Errorf("Upload Error")
	}
	n, err := s3Client.Write(data)
	assert.EqualError(t, err, "Upload Error")
	assert.Equal(t, n, 0)
}

func TestRead(t *testing.T) {
	c := &mockS3{}
	s3Client := S3{downloader: c,
		bucket: "bucket",
		key:    "key",
	}

	data := make([]byte, 5)
	// success downloading
	c.DownloadFunc = func(w io.WriterAt, in *s3.GetObjectInput, opts ...func(*s3manager.Downloader)) (int64, error) {
		return int64(len(data)), nil
	}
	_, err := s3Client.Read(data)
	assert.EqualError(t, err, "EOF")

	// error downloading; need to reset
	s3Client.ready = false
	c.DownloadFunc = func(w io.WriterAt, in *s3.GetObjectInput, opts ...func(*s3manager.Downloader)) (int64, error) {
		return 0, fmt.Errorf("Download Error")
	}
	_, err = s3Client.Read(data)
	assert.EqualError(t, err, "Download Error")
}
