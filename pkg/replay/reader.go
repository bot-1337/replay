package replay

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/sirupsen/logrus"
)

type readerFunc func(path string) (output string, found bool, err error)

func localReader(path string) (output string, found bool, err error) {
	fileObject, err := os.Open(path)
	if err != nil {
		err = fmt.Errorf("error reading file (%s): %w", path, err)
		return "", false, err
	}
	defer fileObject.Close()

	gzReader, err := gzip.NewReader(fileObject)
	if err != nil {
		err = fmt.Errorf("error setting up gzip reader for file (%s): %w", path, err)
		return "", false, err
	}

	buffer := new(bytes.Buffer)
	buffer.ReadFrom(gzReader)
	output = output + buffer.String()

	logrus.Debug(output)

	return output, true, nil
}

func s3Reader(path string) (output string, found bool, err error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"), // the bucket is in us-east-1
	})
	if err != nil {
		err = fmt.Errorf("setting up aws session: %w", err)
		return "", false, err
	}
	svc := s3.New(sess)

	// format s3 inputs
	path = strings.Replace(path, "s3://", "", 1)
	pathSplit := strings.Split(path, "/")
	if len(pathSplit) < 2 {
		err = fmt.Errorf("the file (%s) path was invalid", path)
		return "", false, err
	}
	bucket := pathSplit[0]
	key := strings.Join(pathSplit[1:], "/")

	// get from s3
	result, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		err = fmt.Errorf("error with s3 GetObject for path (%s): %w", path, err)
		return "", false, err
	}

	gzReader, err := gzip.NewReader(result.Body)
	if err != nil {
		err = fmt.Errorf("error setting up gzip reader for file (%s): %w", path, err)
		return "", false, err
	}

	buffer := new(bytes.Buffer)
	buffer.ReadFrom(gzReader)
	output = output + buffer.String()

	logrus.Debug(output)

	// TODO: cache to local filesystem

	return output, true, nil
}
