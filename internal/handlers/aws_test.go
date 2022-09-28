package handlers

import (
	"github.com/kubefirst/kubefirst/internal/services"
	"github.com/kubefirst/kubefirst/pkg"
	"os"
	"testing"
)

func TestAwsHandler_UploadFile(t *testing.T) {

	// mock S3 client
	s3Client := pkg.MockS3{}
	awsService := services.AwsService{
		S3Client: s3Client,
	}
	awsHandler := AwsHandler{
		Service: awsService,
	}

	tmpFile, err := os.CreateTemp("", "tmp-file.ext")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(tmpFile.Name())

	err = awsHandler.UploadFile("bucket-name", tmpFile, "", tmpFile.Name())
	if err != nil {
		t.Error(err)
	}

	tmpFile.Close()
}
