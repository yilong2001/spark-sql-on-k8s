/*
Copyright 2018 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/go-cloud/blob/s3blob"
	//"gocloud.dev/blob/s3blob"

    "github.com/minio/minio-go/v7"
    miniocred "github.com/minio/minio-go/v7/pkg/credentials")

type blobS3 struct {
	s *session.Session
}

func (blob blobS3) setPublicACL(
	ctx context.Context,
	bucket string,
	filePath string) error {
	acl := "public-read"
	svc := s3.New(blob.s)

	if _, err := svc.PutObjectAcl(&s3.PutObjectAclInput{Bucket: &bucket, Key: &filePath, ACL: &acl}); err != nil {
		return fmt.Errorf("failed to set ACL on S3 object %s: %v", filePath, err)
	}

	return nil
}

func newS3Blob(
	ctx context.Context,
	bucket string,
	endpoint string,
	region string) (*uploadHandler, error) {
	// AWS SDK does require specifying regions, thus set it to default S3 region
	if region == "" {
		region = "us-east1"
	}
	c := &aws.Config{
		Region:   aws.String(region),
		Endpoint: aws.String(endpoint),
	}
	sess := session.Must(session.NewSession(c))
	b, err := s3blob.OpenBucket(ctx, sess, bucket)
	return &uploadHandler{
		blob:             blobS3{s: sess},
		ctx:              ctx,
		b:                b,
		blobUploadBucket: bucket,
		blobEndpoint:     endpoint,
		hdpScheme:        "s3a",
	}, err
}

func newPrivateS3Blob(
	ctx context.Context,
	bucket string,
	endpoint string,
	accesskey string,
	secritkey string,
	region string) (*uploadHandler, error) {
	// AWS SDK does require specifying regions, thus set it to default S3 region
	if region == "" {
		region = "us-east1"
	}

	c := &aws.Config{
		Region:   aws.String(region),
		Endpoint: aws.String(endpoint),
		Credentials: credentials.NewStaticCredentials(accesskey, secritkey, ""),
		DisableSSL: aws.Bool(true),
	}

	sess := session.Must(session.NewSession(c))
	b, err := s3blob.OpenBucket(ctx, sess, "/"+bucket)
	return &uploadHandler{
		blob:             blobS3{s: sess},
		ctx:              ctx,
		b:                b,
		blobUploadBucket: bucket,
		blobEndpoint:     endpoint,
		hdpScheme:        "s3a",
	}, err
}

type minioHandler struct {
	clinet *minio.Client
}

func newMinioHandler(
	ctx context.Context,
	bucket string,
	endpoint string,
	accesskey string,
	secritkey string,
	region string) (*minioHandler, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
        Creds:  miniocred.NewStaticV4(accesskey, secritkey, ""),
        Secure: false,
    })
    if err != nil {
        log.Error(err)
        return nil, err
    }

    err = minioClient.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
    if err != nil {
    	log.Error(err)
        // 检查存储桶是否已经存在。
        exists, errBucketExists := minioClient.BucketExists(ctx, bucket)
        if errBucketExists == nil && exists {
            log.Errorf("bucket already exist : %s\n", bucket)
        } else {
            log.Error(err)
        }
    } else {
        log.Infof("Successfully created bucket : %s\n", bucket)
    }

    return &minioHandler{minioClient}, nil
}
