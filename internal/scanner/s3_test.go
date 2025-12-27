package scanner

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type MockS3Client struct {
	Output *s3.ListObjectsV2Output
	Err    error
}

func (m *MockS3Client) ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	return m.Output, m.Err
}

func TestScanS3(t *testing.T) {
	mockClient := &MockS3Client{
		Output: &s3.ListObjectsV2Output{
			Contents: []types.Object{
				{Key: aws.String("folder/image.jpg")},
				{Key: aws.String("folder/doc.txt")},
				{Key: aws.String("folder/photo.PNG")},
			},
		},
	}

	scanner := &S3Scanner{
		Client: mockClient,
		Bucket: "test-bucket",
	}

	files, err := scanner.Scan(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(files))
	}
	
	expected := []string{"folder/image.jpg", "folder/photo.PNG"}
	for i, f := range files {
		if f != expected[i] {
			t.Errorf("Index %d: expected %s, got %s", i, expected[i], f)
		}
	}
}
