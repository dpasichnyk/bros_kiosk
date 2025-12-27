package scanner

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3ClientAPI interface {
	ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
}

type S3Scanner struct {
	Client S3ClientAPI
	Bucket string
	Prefix string
}

func NewS3Scanner(client S3ClientAPI, bucket, prefix string) *S3Scanner {
	return &S3Scanner{
		Client: client,
		Bucket: bucket,
		Prefix: prefix,
	}
}

func (s *S3Scanner) Scan(ctx context.Context) ([]string, error) {
	var files []string
	paginator := s3.NewListObjectsV2Paginator(s.Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.Bucket),
		Prefix: aws.String(s.Prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, obj := range page.Contents {
			if obj.Key == nil {
				continue
			}
			key := *obj.Key
			ext := strings.ToLower(filepath.Ext(key))
			if SupportedExts[ext] {
				files = append(files, key)
			}
		}
	}

	return files, nil
}
