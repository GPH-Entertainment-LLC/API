package core

import (
	"fmt"
	"mime/multipart"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func UploadImage(file *multipart.FileHeader, key string, bucket string) error {
	fmt.Println("Inside upload")
	region := os.Getenv("REGION")

	// Create a new AWS session
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))

	// Initialize S3 service
	s3Svc := s3.New(sess)

	// Open the file
	fileContent, err := file.Open()
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("Getting file content")
	defer fileContent.Close()

	// Upload the file to S3
	uploadInput := &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(key),
		Body:          fileContent,
		ContentType:   aws.String(file.Header.Get("Content-Type")),
		ContentLength: aws.Int64(file.Size),
	}

	_, err = s3Svc.PutObject(uploadInput)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
