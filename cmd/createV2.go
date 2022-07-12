/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"log"
	"time"
	"github.com/kubefirst/nebulous/internal/progressPrinter"
	"github.com/spf13/cobra"
	"github.com/kubefirst/nebulous/configs"
	"github.com/kubefirst/nebulous/pkg"
)

// createV2Cmd represents the createV2 command


var createV2Cmd = &cobra.Command{
	Use:   "createV2",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("createV2 called")
		config := configs.ReadConfig()

		skipVault, err := cmd.Flags().GetBool("skip-vault")
		if err != nil {
			log.Panic(err)
		}
		skipGitlab, err := cmd.Flags().GetBool("skip-gitlab")
		if err != nil {
			log.Panic(err)
		}
		dryRun, err := cmd.Flags().GetBool("dry-run")
		if err != nil {
			log.Panic(err)
		}

		infoCmd.Run(cmd, args)
				
	
		progressPrinter.AddTracker("step-0", "Load Configs", 1)
		progressPrinter.IncrementTracker("step-0", 1)
	

		time.Sleep(time.Millisecond * 500)
		//! ---
		fakeApplyBase()
		//! ---

		//Once bease is ready, deploy Informer. 

		fakeApplyVault()

		fakeApplyGitlab()

		destroyAll()

		if !dryRun {
			log.Printf("Not Dry-run %s,%s,%s",skipVault, skipGitlab, config )
		} else {
			log.Printf("[#99] Dry-run mode")
		}
		time.Sleep(time.Millisecond * 2000)
	},
}

func init() {
	rootCmd.AddCommand(createV2Cmd)
	// todo: make this an optional switch and check for it or viper
	createV2Cmd.Flags().Bool("destroy", false, "destroy resources")
	createV2Cmd.Flags().Bool("dry-run", false, "set to dry-run mode, no changes done on cloud provider selected")
	createV2Cmd.Flags().Bool("skip-gitlab", false, "Skip GitLab lab install and vault setup")
	createV2Cmd.Flags().Bool("skip-vault", false, "Skip post-gitClient lab install and vault setup")

	progressPrinter.GetInstance()
	progressPrinter.SetupProgress(5)
}


func fakeApplyBase(){
	config := configs.ReadConfig()
	progressPrinter.AddTracker("step-base", "Apply Base ", 3)
	progressPrinter.LogMessage("- Waiting bootstrap")
	progressPrinter.IncrementTracker("step-base", 1)
	time.Sleep(time.Millisecond * 500)
	progressPrinter.LogMessage("- Sleeping")
	progressPrinter.IncrementTracker("step-base", 1)
	progressPrinter.LogMessage("- Execute deployment")
	time.Sleep(time.Millisecond * 500)
	_, _, err := pkg.ExecShellReturnStrings(config.KubectlClientPath, "apply", "-f", fmt.Sprintf("./yaml/mock-argo-server.yaml"))
	if err != nil {
		log.Panicf("failed to call execute kubectl apply of argocd patch to adopt gitlab: %s", err)
	}
	progressPrinter.LogMessage("- EKS Cluster Ready")
	time.Sleep(time.Millisecond * 500)
	progressPrinter.IncrementTracker("step-base", 1)

}

func fakeApplyVault(){
	config := configs.ReadConfig()
	progressPrinter.AddTracker("step-vault", "Apply Vault ", 3)
	progressPrinter.LogMessage("- Waiting bootstrap")
	progressPrinter.IncrementTracker("step-vault", 1)
	time.Sleep(time.Millisecond * 500)
	progressPrinter.LogMessage("- Sleeping")
	progressPrinter.IncrementTracker("step-vault", 1)
	progressPrinter.LogMessage("- Execute deployment")
	time.Sleep(time.Millisecond * 500)
	_, _, err := pkg.ExecShellReturnStrings(config.KubectlClientPath, "apply", "-f", fmt.Sprintf("./yaml/mock-vault-pod.yaml"))
	if err != nil {
		log.Panicf("failed to call execute kubectl apply of argocd patch to adopt gitlab: %s", err)
	}
	progressPrinter.IncrementTracker("step-vault", 1)
	
}

func fakeApplyGitlab(){
	config := configs.ReadConfig()
	progressPrinter.AddTracker("step-gitlab", "Apply Gitlab ", 3)
	progressPrinter.LogMessage("- Waiting bootstrap")
	progressPrinter.IncrementTracker("step-gitlab", 1)
	time.Sleep(time.Millisecond * 500)
	progressPrinter.LogMessage("- Sleeping")
	progressPrinter.IncrementTracker("step-gitlab", 1)
	progressPrinter.LogMessage("- Execute deployment")
	time.Sleep(time.Millisecond * 500)
	_, _, err := pkg.ExecShellReturnStrings(config.KubectlClientPath, "apply", "-f", fmt.Sprintf("./yaml/mock-gitlab.yaml"))
	if err != nil {
		log.Panicf("failed to call execute kubectl apply of argocd patch to adopt gitlab: %s", err)
	}
	progressPrinter.IncrementTracker("step-gitlab", 1)

	
}


func destroyAll(){
	config := configs.ReadConfig()
	progressPrinter.AddTracker("step-delete", "Destroy All ", 3)
	_, _, err := pkg.ExecShellReturnStrings(config.KubectlClientPath, "delete", "-f", fmt.Sprintf("./yaml/mock-gitlab.yaml"))
	if err != nil {
		log.Panicf("failed to call execute kubectl apply of argocd patch to adopt gitlab: %s", err)
	}
	progressPrinter.LogMessage("- Destroyed Gitlab")
	time.Sleep(time.Millisecond * 500)
	progressPrinter.IncrementTracker("step-delete", 1)

	_, _, err = pkg.ExecShellReturnStrings(config.KubectlClientPath, "delete", "-f", fmt.Sprintf("./yaml/mock-vault-pod.yaml"))
	if err != nil {
		log.Panicf("failed to call execute kubectl apply of argocd patch to adopt gitlab: %s", err)
	}
	progressPrinter.LogMessage("- Destroyed Vault")
	time.Sleep(time.Millisecond * 500)
	progressPrinter.IncrementTracker("step-delete", 1)

	_, _, err = pkg.ExecShellReturnStrings(config.KubectlClientPath, "delete", "-f", fmt.Sprintf("./yaml/mock-argo-server.yaml"))
	if err != nil {
		log.Panicf("failed to call execute kubectl apply of argocd patch to adopt gitlab: %s", err)
	}
	progressPrinter.LogMessage("- Destroyed Argo")
	time.Sleep(time.Millisecond * 500)
	progressPrinter.IncrementTracker("step-delete", 1)

	
}
