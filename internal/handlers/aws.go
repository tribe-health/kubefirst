package handlers

import (
	"fmt"
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

func (handler AwsHandler) UploadFolder(baseFolder string, prefix string, bucketName string) error {

	walker := make(chan string)

	// todo: send it to a function
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

		relativeFilePath, err := filepath.Rel(baseFolder, baseFolder+path)
		if err != nil {
			return err
		}

		// os.Stat ?  todo:
		file, err := os.Open(baseFolder + path)
		if err != nil {
			return err
		}
		defer file.Close()

		err = handler.Service.NewUploadFile(bucketName, prefix, baseFolder+path, relativeFilePath)
		if err != nil {
			return err
		}

		fmt.Println("upload ok!...")
	}

	return nil
}

func (handler AwsHandler) UploadFile(bucketName string, prefix string, remoteFilename string, localFilename string) error {

	err := handler.Service.NewUploadFile(bucketName, prefix, remoteFilename, localFilename)
	if err != nil {
		return err
	}

	return nil
}
