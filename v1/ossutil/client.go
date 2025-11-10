package ossutil

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// type

type OssClient struct {
	seClient *s3.Client
}

func NewOssClient(region string) *OssClient {
	return &OssClient{
		seClient: s3.NewFromConfig(aws.Config{
			Region: region,
		}),
	}
}
