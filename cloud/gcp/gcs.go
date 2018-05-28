package gcp

import (
	"context"

	"cloud.google.com/go/storage"
)

// GCS is Google Cloud Storage client
type GCS struct {
	client    *storage.Client
	bucket    string
	key       string
	readReady bool
	reader    *storage.Reader
}

// NewGCS creates new GCS client
func NewGCS(bucket, key string) (*GCS, error) {
	client, err := storage.NewClient(context.Background())
	if err != nil {
		return nil, err
	}

	return &GCS{
		client:    client,
		bucket:    bucket,
		key:       key,
		readReady: false,
	}, nil
}

// Write writes data to GCS bucket
func (s *GCS) Write(data []byte) (int, error) {
	ctx := context.Background()
	w := s.client.Bucket(s.bucket).Object(s.key).NewWriter(ctx)
	defer w.Close()

	return w.Write(data)
}

// Read reads data from GCS bucket
func (s *GCS) Read(data []byte) (int, error) {
	if !s.readReady {
		var err error
		ctx := context.Background()
		s.reader, err = s.client.Bucket(s.bucket).Object(s.key).NewReader(ctx)
		if err != nil {
			return 0, err
		}
		s.readReady = true
	}

	n, err := s.reader.Read(data)
	if err != nil {
		s.readReady = false
	}

	return n, err
}
