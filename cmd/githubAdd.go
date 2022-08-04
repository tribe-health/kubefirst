/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"log"

	"github.com/kubefirst/kubefirst/internal/githubWrapper"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// githubAddCmd represents the setupGithub command
var githubAddCmd = &cobra.Command{
	Use:   "add-github",
	Short: "Setup github for kubefirst install",
	Long:  `TBD`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("githubAddCmd called")
		org, err := cmd.Flags().GetString("github-org")
		if err != nil {
			return err
		}
		log.Println("Org used:", org)
		dryrun, err := cmd.Flags().GetBool("dry-run")
		if err != nil {
			return err
		}
		log.Println("dry-run:", dryrun)

		if viper.GetBool("github.repo.added") {
			log.Println("github.repo.added already executed, skiped")
			return nil
		}
		if dryrun {
			log.Printf("[#99] Dry-run mode, githubAddCmd skipped.")
			return nil
		}

		gitWrapper := githubWrapper.New()
		gitWrapper.CreatePrivateRepo(org, "gitops", "Kubefirst Gitops")
		gitWrapper.CreatePrivateRepo(org, "metaphor", "Sample Kubefirst App")

		viper.Set("github.repo.added", true)
		viper.WriteConfig()
		return nil
	},
}

func init() {
	actionCmd.AddCommand(githubAddCmd)
	currentCommand := githubAddCmd
	currentCommand.Flags().Bool("dry-run", false, "set to dry-run mode, no changes done on cloud provider selected")
	currentCommand.Flags().String("github-org", "", "Github Org of repos")
	viper.BindPFlag("github.org", githubAddCmd.Flags().Lookup("github-org"))

}
