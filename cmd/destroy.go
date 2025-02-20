package cmd

import (
	"fmt"
	"github.com/kubefirst/kubefirst/pkg"
	"log"
	"time"

	"github.com/kubefirst/kubefirst/internal/flagset"
	"github.com/kubefirst/kubefirst/internal/gitClient"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// destroyCmd represents the destroy command
var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "destroy Kubefirst management cluster",
	Long:  "destroy all the resources installed via Kubefirst installer",
	RunE: func(cmd *cobra.Command, args []string) error {

		//Destroy is implemented based on the flavor selected.
		if viper.GetString("cloud") == pkg.CloudK3d {
			err := destroyLocalGithubCmd.RunE(cmd, args)
			if err != nil {
				log.Println("Erroring destroying local+github:", err)
				return err
			}
		} else if viper.GetString("cloud") == flagset.CloudAws {
			if viper.GetString("gitprovider") == gitClient.Github {
				err := destroyAwsGithubCmd.RunE(cmd, args)
				if err != nil {
					log.Println("Error destroying aws+github:", err)
					return err
				}

			} else if viper.GetString("gitprovider") == gitClient.Gitlab {
				err := destroyAwsGitlabCmd.RunE(cmd, args)
				if err != nil {
					log.Println("Error destroying aws+gitlab:", err)
					return err
				}
			} else {
				return fmt.Errorf("not supported git-provider")
			}

		} else {
			return fmt.Errorf("not supported mode")
		}

		log.Println("terraform base destruction complete")
		time.Sleep(time.Millisecond * 100)

		return nil
	},
}

func init() {
	clusterCmd.AddCommand(destroyCmd)
	currentCommand := destroyCmd
	flagset.DefineGlobalFlags(currentCommand)
	flagset.DefineDestroyFlags(currentCommand)
}
