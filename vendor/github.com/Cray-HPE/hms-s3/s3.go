// MIT License
//
// (C) Copyright [2020-2021] Hewlett Packard Enterprise Development LP
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.

package hms_s3

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type ConnectionInfo struct {
	AccessKey string
	SecretKey string
	Endpoint  string
	Bucket    string
	Region    string
}

func (obj *ConnectionInfo) Equals(other ConnectionInfo) (equals bool) {
	if obj.Region == other.Region &&
		obj.Bucket == other.Bucket &&
		obj.Endpoint == other.Endpoint &&
		obj.SecretKey == other.SecretKey &&
		obj.AccessKey == other.AccessKey {
		equals = true
	}
	return equals
}

func (obj *ConnectionInfo) Validate() error {
	if obj.AccessKey == "" {
		return errors.New("s3 access key is empty")
	}
	if obj.SecretKey == "" {
		return errors.New("s3 secret key is empty")
	}
	if obj.Endpoint == "" {
		return errors.New("s3 endpoint is empty")
	}
	if obj.Bucket == "" {
		return errors.New("s3 bucket is empty")
	}
	if obj.Region == "" {
		return errors.New("s3 region is empty")
	}
	return nil
}

func NewConnectionInfo(AccessKey string, SecretKey string, Endpoint string, Bucket string,
	Region string) (c ConnectionInfo) {
	c = ConnectionInfo{
		AccessKey: AccessKey,
		SecretKey: SecretKey,
		Endpoint:  Endpoint,
		Bucket:    Bucket,
		Region:    Region,
	}
	return c
}

func LoadConnectionInfoFromEnvVars() (info ConnectionInfo, err error) {
	// There is no static default for the access and secret keys.
	info.AccessKey = os.Getenv("S3_ACCESS_KEY")
	if info.AccessKey == "" {
		err = errors.New("access key cannot be empty")
	}
	info.SecretKey = os.Getenv("S3_SECRET_KEY")
	if info.SecretKey == "" {
		err = errors.New("secret key cannot be empty")
	}
	info.Endpoint = os.Getenv("S3_ENDPOINT")
	if info.Endpoint == "" {
		err = errors.New("endpoint cannot be empty")
	}
	info.Bucket = os.Getenv("S3_BUCKET")
	if info.Bucket == "" {
		info.Bucket = "default"
	}
	// Default to "" if there is no region defined in the environment.
	info.Region = os.Getenv("S3_REGION")
	if info.Region == "" {
		info.Region = "default"
	}
	return info, err
}

type S3Client struct {
	Session  *session.Session
	Uploader *s3manager.Uploader
	S3       *s3.S3
	//Service service
	ConnInfo ConnectionInfo
}

// NewS3Client only sets up the connection to S3, it *does not* test that connection. For that, call PingBucket().
func NewS3Client(info ConnectionInfo, httpClient *http.Client) (*S3Client, error) {
	var client S3Client
	var err error

	client.Session, err = session.NewSession(aws.NewConfig().
		WithCredentials(credentials.NewStaticCredentials(info.AccessKey, info.SecretKey, "")).
		WithEndpoint(info.Endpoint).
		WithHTTPClient(httpClient).
		WithRegion(info.Region).
		WithS3ForcePathStyle(true),
	)

	if err != nil {
		return nil, fmt.Errorf("failed setting up new session: %w", err)
	}

	if httpClient != nil {
		client.Session.Config.HTTPClient = httpClient
	}

	client.S3 = s3.New(client.Session)
	client.ConnInfo = info

	client.Uploader = s3manager.NewUploader(client.Session)

	return &client, nil
}

func GetCreateBucketInputWithACL(bucketName string, acl string) *s3.CreateBucketInput {
	return &s3.CreateBucketInput{
		ACL:    aws.String(acl),
		Bucket: aws.String(bucketName),
	}
}

// CreateBucketWithACL creates a new bucket in S3 with the provided ACL.
func (client *S3Client) CreateBucketWithACL(bucketName string, acl string) (*s3.CreateBucketOutput, error) {
	return client.S3.CreateBucket(GetCreateBucketInputWithACL(bucketName, acl))
}

// PingBucket will test the connection to S3 a single time. If you're using this as a measure for whether S3 is
// responsive, call it in a loop looking for nil err returned.
func (client *S3Client) PingBucket() error {
	// Test connection to S3
	_, err := client.S3.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(client.ConnInfo.Bucket),
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			switch awsErr.Code() {
			case s3.ErrCodeNoSuchBucket:
				err := fmt.Errorf("bucket %s does not exist at %s",
					client.ConnInfo.Bucket, client.ConnInfo.Endpoint)
				return err
			}
			err := fmt.Errorf("encountered error during head_bucket operation for bucket %s at %s: %w",
				client.ConnInfo.Bucket, client.ConnInfo.Endpoint, err)
			return err
		}
	}

	return nil
}

// SetBucket updates the connection info to use the newly passed in bucket name.
func (client *S3Client) SetBucket(newBucket string) {
	client.ConnInfo.Bucket = newBucket
}

// GET

func (client *S3Client) GetObjectInput(key string) *s3.GetObjectInput {
	return &s3.GetObjectInput{
		Bucket: aws.String(client.ConnInfo.Bucket),
		Key:    aws.String(key),
	}
}

func (client *S3Client) GetObject(key string) (*s3.GetObjectOutput, error) {
	return client.S3.GetObject(client.GetObjectInput(key))
}

func (client *S3Client) GetURL(key string, expire time.Duration) (string, error) {
	req, _ := client.S3.GetObjectRequest(client.GetObjectInput(key))
	urlStr, err := req.Presign(expire)

	return urlStr, err
}

// PUT

func (client *S3Client) PutObjectInputBytes(key string, payloadBytes []byte) *s3.PutObjectInput {
	r := bytes.NewReader(payloadBytes)

	return &s3.PutObjectInput{
		Bucket: aws.String(client.ConnInfo.Bucket),
		Key:    aws.String(key),
		Body:   r,
	}
}

func (client *S3Client) PutObjectInputFile(key string, file *os.File) *s3.PutObjectInput {
	return &s3.PutObjectInput{
		Bucket: aws.String(client.ConnInfo.Bucket),
		Key:    aws.String(key),
		Body:   file,
	}
}

func (client *S3Client) PutObjectInputFileACL(key string, file *os.File, acl string) *s3.PutObjectInput {
	return &s3.PutObjectInput{
		ACL:    aws.String(acl),
		Bucket: aws.String(client.ConnInfo.Bucket),
		Key:    aws.String(key),
		Body:   file,
	}
}

func (client *S3Client) UploadInputACL(key string, file *os.File, acl string) *s3manager.UploadInput {
	return &s3manager.UploadInput{
		ACL:    aws.String(acl),
		Bucket: aws.String(client.ConnInfo.Bucket),
		Key:    aws.String(key),
		Body:   file,
	}
}

func (client *S3Client) PutObject(key string, payloadBytes []byte) (*s3.PutObjectOutput, error) {
	return client.S3.PutObject(client.PutObjectInputBytes(key, payloadBytes))
}

func (client *S3Client) PutFile(key string, file *os.File) (*s3.PutObjectOutput, error) {
	return client.S3.PutObject(client.PutObjectInputFile(key, file))
}

func (client *S3Client) PutFileWithACL(key string, file *os.File, acl string) (*s3.PutObjectOutput, error) {
	return client.S3.PutObject(client.PutObjectInputFileACL(key, file, acl))
}

func (client *S3Client) UploadFileWithACL(key string, file *os.File, acl string) (*s3manager.UploadOutput, error) {
	return client.Uploader.Upload(client.UploadInputACL(key, file, acl))
}

// Delete
func (client *S3Client) DeleteObjectInput(key string) *s3.DeleteObjectInput {
	return &s3.DeleteObjectInput{
		Bucket: aws.String(client.ConnInfo.Bucket),
		Key:    aws.String(key),
	}
}

func (client *S3Client) DeleteObject(key string) (*s3.DeleteObjectOutput, error) {
	return client.S3.DeleteObject(client.DeleteObjectInput(key))
}
