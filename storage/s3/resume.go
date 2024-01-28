package s3

import (
	"context"
	"io"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"multi-tenant-HR-information-system-backend/httperror"
)

func (s *S3) UploadResume(file io.Reader, jobApplicationId string, firstName string, lastName string, fileExt string) (url string, err error) {
	key := fmt.Sprintf("job-applications/%s/%s_%s_resume%s", jobApplicationId, firstName, lastName, fileExt)

	_, err = s.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(BucketName),
		Key: aws.String(key),
		Body: file,
	})
	if err != nil {
		return "", httperror.NewInternalServerError(err)
	}

	url = fmt.Sprintf("%s/%s/%s", s.baseUrl, BucketName, key)
	log.Println(url)

	return url, nil
}
