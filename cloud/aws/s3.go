package aws

import (
	"bytes"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// S3 is AWS S3 client
type S3 struct {
	uploader interface {
		Upload(*s3manager.UploadInput, ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)
	}
	downloader interface {
		Download(io.WriterAt, *s3.GetObjectInput, ...func(*s3manager.Downloader)) (int64, error)
	}
	bucket    string
	key       string
	readReady bool
	bufReader *bytes.Buffer
}

// NewS3WithSession creates new AWS S3 client with session sess
func NewS3WithSession(bucket, key string, sess *session.Session) (*S3, error) {
	uploader := s3manager.NewUploader(sess)
	downloader := s3manager.NewDownloader(sess)

	return &S3{
		uploader:   uploader,
		downloader: downloader,
		bucket:     bucket,
		key:        key,
		readReady:  false,
	}, nil
}

// NewS3 returns new AWS S3 client
func NewS3(bucket, key string) (*S3, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	return NewS3WithSession(bucket, key, sess)
}

// Write writes data to S3 bucket
func (s *S3) Write(data []byte) (int, error) {
	body := bytes.NewBuffer(data)
	// Upload the file to S3.
	_, err := s.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.key),
		Body:   body,
	})
	if err != nil {
		return 0, err
	}

	return len(data), nil
}

// Read reads data from S3 bucket
func (s *S3) Read(data []byte) (int, error) {
	if !s.readReady {
		buf := &aws.WriteAtBuffer{}
		// Write the contents of S3 Object to the file
		_, err := s.downloader.Download(buf, &s3.GetObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(s.key),
		})
		if err != nil {
			return 0, err
		}
		s.readReady = true
		s.bufReader = bytes.NewBuffer(buf.Bytes())
	}

	return s.bufReader.Read(data)
}
