/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/kubefirst/kubefirst/configs"
	"github.com/kubefirst/kubefirst/internal/domain"
	"github.com/kubefirst/kubefirst/internal/downloadManager"
	"github.com/kubefirst/kubefirst/internal/github"
	"github.com/kubefirst/kubefirst/internal/handlers"
	"github.com/kubefirst/kubefirst/internal/progressPrinter"
	"github.com/kubefirst/kubefirst/internal/repo"
	"github.com/kubefirst/kubefirst/internal/services"
	"github.com/kubefirst/kubefirst/pkg"
	"github.com/segmentio/analytics-go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// localCmd represents the local command
var localCmd = &cobra.Command{
	Use:   "local",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Println("\nwelcome to the local kubefirst platform installation!")

		infoCmd.Run(cmd, args)
		config := configs.ReadConfig()

		if err := pkg.ValidateK1Folder(config.K1FolderPath); err != nil {
			return err
		}

		progressPrinter.AddTracker("step-download", pkg.DownloadDependencies, 3)
		progressPrinter.AddTracker("step-gitops", pkg.CloneAndDetokenizeGitOpsTemplate, 1)
		progressPrinter.AddTracker("step-ssh", pkg.CreateSSHKey, 1)
		progressPrinter.AddTracker("step-telemetry", pkg.SendTelemetry, 1)
		hackSilent := false
		progressPrinter.SetupProgress(progressPrinter.TotalOfTrackers(), hackSilent)

		// //* command line flags - todo process these in another function with validation
		sendTelemetryFlag := false
		// sendTelemetryFlag, err := cmd.Flags().GetBool("send-telemetry")
		// if err != nil {
		// 	return err
		// }
		log.Print(fmt.Sprintf("\n\n %b \n\n", sendTelemetryFlag))
		viper.Set("kubefirst.telemetry", sendTelemetryFlag)
		viper.WriteConfig()
		// cloudFlag, err := cmd.Flags().GetString("cloud")
		// if err != nil {
		// 	return err
		// }

		atlantisWebhookSecret := pkg.Random(20)
		viper.Set("github.atlantis.webhook.secret", atlantisWebhookSecret)

		//! no viper.Write() in other files, only allowed in local.go and we must return values we need to this context
		if os.Getenv("GITHUB_AUTH_TOKEN") != "" {

			githubAuthToken := os.Getenv("GITHUB_AUTH_TOKEN")
			viper.Set("github.token", githubAuthToken)
			viper.WriteConfig()

			log.Println("getting github authenicated user")
			hackUrl := "http://localhost:3000/githubUser"
			githubOwner := github.GetGithubOwner(githubAuthToken, hackUrl)
			githubOwnerGitopsUrl := fmt.Sprintf("https://github.com/%s/gitops", githubOwner)
			viper.Set("github.owner", githubOwner)
			viper.Set("github.user", githubOwner)
			viper.Set("github.url.gitops", githubOwnerGitopsUrl)
			viper.WriteConfig()

		} else {
			log.Fatal("error: cannot create a local cluster without a github auth token.\n  please `export GITHUB_AUTH_TOKEN=$YOUR_GITHUB_PERSONAL_ACCESS_TOKEN` in your terminal.")
		}

		// send telemetry

		if sendTelemetryFlag {
			domainOrMachineId := "MACHINE_ID_BECAUSE_LOCAL" // todo
			sendTelemetry(pkg.MetricInitStarted, domainOrMachineId)
		}

		// //! tracker 0
		log.Println("installing kubefirst dependencies")
		progressPrinter.IncrementTracker("step-download", 1)
		err := downloadManager.DownloadTools(config)
		if err != nil {
			return err
		}
		log.Println("dependency installation complete")
		progressPrinter.IncrementTracker("step-download", 1)
		err = downloadManager.DownloadLocalTools(config)
		if err != nil {
			return err
		}

		//Fix incomplete bar, please don't remove it.
		progressPrinter.IncrementTracker("step-download", 1)

		log.Println("creating an ssh key pair for your new cloud infrastructure")
		pkg.CreateSshKeyPair()
		log.Println("ssh key pair creation complete")
		progressPrinter.IncrementTracker("step-ssh", 1)

		repo.PrepareKubefirstTemplateRepo(hackSilent, config, viper.GetString("gitops.owner"), viper.GetString("gitops.repo"), viper.GetString("gitops.branch"), viper.GetString("template.tag"))
		log.Println("clone and detokenization of gitops-template repository complete")
		progressPrinter.IncrementTracker("step-gitops", 1)

		log.Println("sending init completed metric")

		if sendTelemetryFlag {
			domainOrMachineId := "MACHINE_ID_BECAUSE_LOCAL" // todo
			sendTelemetry(pkg.MetricInitCompleted, domainOrMachineId)
		}

		//! tracker 8
		progressPrinter.IncrementTracker("step-telemetry", 1)
		time.Sleep(time.Millisecond * 100)

		// informUser("init is done!\n", globalFlags.SilentMode)
		// finish init

		// create import

		// createk3d

		// start
		os.Exit(0)
		return nil // todo return error

	},
}

//*
//* init here
//*
func init() {
	rootCmd.AddCommand(localCmd)

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	localCmd.Flags().Bool("send-telemetry", false, "whether or not to send telemetry")
	localCmd.Flags().String("cloud", "k3d", "the cloud to install the kubefirst platform to")
}

//!
//!
//! functions below
//!
//!
func sendTelemetry(metric, domainOrMachineId string) {
	var telemetryHandler handlers.TelemetryHandler

	// Instantiates a SegmentIO client to use send messages to the segment API.
	segmentIOClient := analytics.New(pkg.SegmentIOWriteKey)

	// SegmentIO library works with queue that is based on timing, we explicit close the http client connection
	// to force flush in case there is still some pending message in the SegmentIO library queue.
	defer func(segmentIOClient analytics.Client) {
		err := segmentIOClient.Close()
		if err != nil {
			log.Println(err)
		}
	}(segmentIOClient)

	// validate telemetryDomain data
	telemetryDomain, err := domain.NewTelemetry(
		metric,
		domainOrMachineId,
		configs.K1Version,
	)
	if err != nil {
		log.Println(err)
	}
	telemetryService := services.NewSegmentIoService(segmentIOClient)
	telemetryHandler = handlers.NewTelemetryHandler(telemetryService)

	err = telemetryHandler.SendCountMetric(telemetryDomain)
	if err != nil {
		log.Println(err)

	}
}
