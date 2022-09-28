package handlers

import (
	"github.com/kubefirst/kubefirst/internal/flagset"
	"github.com/kubefirst/kubefirst/internal/services"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

// AwsHandler provides base data for Aws Handler methods.
type AwsHandler struct {
	Service  services.AwsService
	CLIFlags *flagset.DestroyFlags
}

// NewAwsHandler creates a new Aws Handler object.
func NewAwsHandler(awsService services.AwsService, cliFlags *flagset.DestroyFlags) AwsHandler {
	return AwsHandler{
		Service:  awsService,
		CLIFlags: cliFlags,
	}
}

// HostedZoneDelete deletes Hosted Zone data based on CLI flags. There are two possibilities to this handler, completely
// delete a hosted zone, or delete all hosted zone records except the base ones (SOA, NS and TXT liveness).
func (handler AwsHandler) HostedZoneDelete(hostedZone string) error {

	// get hosted zone id
	hostedZoneId, err := services.Route53GetHostedZoneId(hostedZone)
	if err != nil {
		return err
	}

	// TXT records
	txtRecords, err := services.Route53ListTXTRecords(hostedZoneId)
	if err != nil {
		return err
	}
	err = services.Route53DeleteTXTRecords(
		hostedZoneId,
		hostedZone,
		handler.CLIFlags.HostedZoneKeepBase,
		txtRecords,
	)
	if err != nil {
		return err
	}

	// A records
	aRecords, err := services.Route53ListARecords(hostedZoneId)
	if err != nil {
		return err
	}
	err = services.Route53DeleteARecords(hostedZoneId, aRecords)
	if err != nil {
		return err
	}

	// deletes full hosted zone, at this point there is only a SOA and a NS record, and deletion will succeed
	if !handler.CLIFlags.HostedZoneKeepBase {
		err := services.Route53DeleteHostedZone(hostedZoneId, hostedZone)
		if err != nil {
			return err
		}
	}
	return nil
}

// UploadFolder receives a folder path, and upload to a service. Parameters are:
//   - baseFolder local folder path
//   - prefix destiny folder to host the upload base files
//   - bucketName name of the destination bucket
func (handler AwsHandler) UploadFolder(baseFolder string, prefix string, bucketName string) error {

	walker := make(chan string)

	go func() {
		fileSystem := os.DirFS(baseFolder)
		err := fs.WalkDir(fileSystem, ".", func(fullFilePath string, folderOrFile fs.DirEntry, err error) error {
			if !folderOrFile.IsDir() {
				walker <- fullFilePath
			}
			return nil
		})
		if err != nil {
			log.Println(err)
		}
		close(walker)
	}()

	for path := range walker {

		// open file to be transmitted
		file, err := os.Open(baseFolder + path)
		if err != nil {
			return err
		}

		fullFilePath := baseFolder + path
		relativeFilePath, err := filepath.Rel(baseFolder, fullFilePath)
		if err != nil {
			return err
		}

		err = handler.Service.NewUploadFile(bucketName, file, prefix, relativeFilePath)
		if err != nil {
			return err
		}

		log.Printf("file %q was uploaded to the %q S3 bucket", fullFilePath, bucketName)
	}

	return nil
}

type mockHandler struct{}

// UploadFile intermediate the request, and send the request to the service provided in the constructor NewAwsHandler.
// The handler doesn't know how to upload a file, it only presents his intention, and let the service do the actual job
// of uploading a file at the service provider provided during the AwsHandler creation.
func (handler AwsHandler) UploadFile(
	bucketName string,
	localFile *os.File,
	prefix string,
	remoteFilename string,
) error {

	err := handler.Service.NewUploadFile(bucketName, localFile, prefix, remoteFilename)
	if err != nil {
		return err
	}

	return nil
}
