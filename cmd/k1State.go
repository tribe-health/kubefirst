package cmd

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/kubefirst/kubefirst/configs"
	"github.com/kubefirst/kubefirst/internal/handlers"
	"github.com/kubefirst/kubefirst/internal/reports"
	"github.com/kubefirst/kubefirst/internal/services"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os"
)

var k1state = &cobra.Command{
	Use:   "state",
	Short: "push and pull Kubefirst configuration to S3 bucket",
	Long:  `Kubefirst configuration can be handed over to another user by pushing the Kubefirst config files to a S3 bucket.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		push, err := cmd.Flags().GetBool("push")
		if err != nil {
			log.Println(err)
		}
		pull, err := cmd.Flags().GetBool("pull")
		if err != nil {
			log.Println(err)
		}

		bucketName, err := cmd.Flags().GetString("bucket-name")
		if err != nil {
			log.Println(err)
		}

		region, err := cmd.Flags().GetString("region")
		if err != nil {
			log.Println(err)
		}

		if !push && !pull {
			fmt.Println(cmd.Help())
			return nil
		}

		if pull && len(region) == 0 {
			return errors.New("region is required when pulling Kubefirst config, please add --region <region-name>")
		}

		// AWS configuration request
		awsConfig, err := services.NewAws()
		if err != nil {
			log.Println(err)
		}

		s3Client := manager.NewUploader(s3.NewFromConfig(awsConfig))
		awsService := services.NewAwsService(s3Client)
		awsHandler := handlers.NewAwsHandler(awsService, nil)

		config := configs.ReadConfig()
		if push {
			// prepare kubefirst file to be uploaded
			kubeFirstConfigFile, err := os.Open(config.KubefirstConfigFilePath)
			if err != nil {
				return err
			}

			// call handler to upload a file
			err = awsHandler.UploadFile(bucketName, kubeFirstConfigFile, "", config.KubefirstConfigFileName)
			if err != nil {
				return err
			}

			// close open file
			err = kubeFirstConfigFile.Close()
			if err != nil {
				return err
			}

			// pass a folder, and let the handler handles the upload process
			err = awsHandler.UploadFolder(config.K1FolderPath, "k1/", bucketName)
			if err != nil {
				return err
			}

			finalMsg := fmt.Sprintf(
				"Kubefirst configuration file %q and %q folder was uploaded to AWS S3 at %q bucket name",
				config.KubefirstConfigFilePath,
				config.K1FolderPath,
				bucketName,
			)

			log.Printf(finalMsg)
			fmt.Println(reports.StyleMessage(finalMsg))
		}

		if pull {

			// at this point user doesn't have kubefirst config file and no aws.region
			viper.Set("aws.region", region)
			if err := viper.WriteConfig(); err != nil {
				return err
			}

			err := services.DownloadS3File(bucketName, config.KubefirstConfigFileName)
			if err != nil {
				return err
			}
			currentFolder, err := os.Getwd()
			finalMsg := fmt.Sprintf("Kubefirst configuration file was downloaded to %q/, and is now available to be copied to %q/",
				currentFolder,
				config.K1FolderPath,
			)

			log.Printf(finalMsg)
			fmt.Println(reports.StyleMessage(finalMsg))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(k1state)

	k1state.Flags().Bool("push", false, "push Kubefirst config file to the S3 bucket")
	k1state.Flags().Bool("pull", false, "pull Kubefirst config file to the S3 bucket")
	k1state.Flags().String("region", "", "set S3 bucket region")
	k1state.Flags().String("bucket-name", "", "set the bucket name to store the Kubefirst config file")
	err := k1state.MarkFlagRequired("bucket-name")
	if err != nil {
		log.Println(err)
		return
	}

}
