package s3

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
    awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"	
)

type S3 struct {
	client *s3.Client
	baseUrl string
}

const BucketName = "hr-information-system"

// Constructs either a real or fake S3 client
func NewS3(credentialsProvider aws.CredentialsProvider, regionCode string) *S3 {
	cfg, _ := config.LoadDefaultConfig(
		context.TODO(),
		config.WithCredentialsProvider(credentialsProvider),
		config.WithRegion(regionCode),		
	)
	client := s3.NewFromConfig(cfg)

	// Check if bucket exists. If not, create it
	_, err := client.HeadBucket(context.TODO(), &s3.HeadBucketInput{
		Bucket: aws.String(BucketName),
	})
	if err != nil {
        var responseError *awshttp.ResponseError
        if errors.As(err, &responseError) && responseError.ResponseError.HTTPStatusCode() == http.StatusNotFound {
			_, err := client.CreateBucket(context.TODO(), &s3.CreateBucketInput{
				Bucket: aws.String(BucketName),
			})
			if err != nil {
				log.Fatalf("Could not instanitate bucket: %s", err)
			}            
        }
	}

	baseUrl := fmt.Sprintf("s3.%s.amazonaws.com", regionCode)

	return &S3{client: client, baseUrl: baseUrl}
}

func NewFakeS3(credentialsProvider aws.CredentialsProvider, fakeUrl string) *S3 {
	cfg, _ := config.LoadDefaultConfig(
		context.TODO(),
		config.WithCredentialsProvider(credentialsProvider),
		config.WithHTTPClient(&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}),
		config.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(func(_, _ string, _ ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: fakeUrl}, nil
			}),
		),
	)

	// Create an Amazon S3 v2 client, important to use o.UsePathStyle
	// alternatively change local DNS settings, e.g., in /etc/hosts
	// to support requests to http://<bucketname>.127.0.0.1:32947/...
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	// Create bucket
	_, err := client.CreateBucket(context.TODO(), &s3.CreateBucketInput{
		Bucket: aws.String(BucketName),
	})
	if err != nil {
		log.Fatalf("Could not instanitate bucket: %s", err)
	}

	return &S3{client: client, baseUrl: fakeUrl}
}
