package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sts"
	"gopkg.in/kyokomi/emoji.v1"
)

type s3Service struct {
	*s3.S3
}

func NewS3Service(profile string, region string, role string) *s3Service {
	var config aws.Config
	if profile != "" && role == "" {
		creds := credentials.NewSharedCredentials("", profile)
		config = aws.Config{Region: aws.String(region),
			Credentials: creds,
			Endpoint:    aws.String(*argEndpoint)}
	} else if profile == "" && role != "" {
		sess := session.Must(session.NewSession())
		creds := stscreds.NewCredentials(sess, role)
		config = aws.Config{Region: aws.String(region),
			Credentials: creds,
			Endpoint:    aws.String(*argEndpoint)}
	} else if profile != "" && role != "" {
		sess := session.Must(session.NewSessionWithOptions(session.Options{Profile: profile}))
		assumeRoler := sts.New(sess)
		creds := stscreds.NewCredentialsWithClient(assumeRoler, role)
		config = aws.Config{Region: aws.String(region),
			Credentials: creds,
			Endpoint:    aws.String(*argEndpoint)}
	} else {
		config = aws.Config{Region: aws.String(region),
			Endpoint: aws.String(*argEndpoint)}
	}
	sess := session.New(&config)
	svc := s3.New(sess)
	sv := &s3Service{
		svc,
	}
	return sv
}

func (sv *s3Service) S3PutObject(keyName, contentType, bucket, cacheControl string, file string) (err error) {
	fileBody, err := os.Open(file)
	if err != nil {
		fmt.Printf("err opening file: %s", err)
	}
	defer fileBody.Close()

	putObj := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(file),
		Body:   fileBody,
	}

	if contentType != "" {
		putObj.ContentType = aws.String(contentType)
	}

	if cacheControl != "" {
		putObj.CacheControl = aws.String(cacheControl)
	}

	// fmt.Println(putObj)
	_, err = sv.PutObject(putObj)

	if err != nil {
		emoji.Println(":bangbang: ソースファイルの S3 へのアップロードに失敗しました.")
	} else {
		emoji.Println(":white_check_mark: ソースファイルの S3 へのアップロードに成功しました.")
	}

	return err
}
