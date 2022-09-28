package pkg

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type MockAWSS3 struct{}

// Upload satisfies S3ClientManager interface
func (myMockS3 MockAWSS3) Upload(ctx context.Context, input *s3.PutObjectInput, opts ...func(*manager.Uploader)) (*manager.UploadOutput, error) {
	return nil, nil
}
