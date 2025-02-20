package flagset

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// GlobalFlags - Global flags
type GlobalFlags struct {
	DryRun       bool
	UseTelemetry bool
	SilentMode   bool
	ConfigFile   string
}

// DefineGlobalFlags - Define global flags
func DefineGlobalFlags(currentCommand *cobra.Command) {
	currentCommand.Flags().Bool("use-telemetry", true, "installer will not send telemetry about this installation")
	currentCommand.Flags().Bool("dry-run", false, "set to dry-run mode, no changes done on cloud provider selected")
	currentCommand.Flags().Bool("silent", false, "enable silent mode will make the UI return less content to the screen")
	currentCommand.Flags().StringP("config", "c", "", "File to be imported to bootstrap configs")
	viper.BindPFlag("config.file", currentCommand.Flags().Lookup("config-load"))
}

// ProcessGlobalFlags - process global flags shared between commands like silent, dry-run and use-telemetry
func ProcessGlobalFlags(cmd *cobra.Command) (GlobalFlags, error) {
	flags := GlobalFlags{}
	config, err := ReadConfigString(cmd, "config")
	if err != nil {
		log.Printf("Error Processing - config flag, error: %v", err)
		return flags, err
	}
	flags.ConfigFile = config
	log.Println("import config source:", flags.ConfigFile)
	if flags.ConfigFile != "" {
		InjectConfigs(flags.ConfigFile)
	}
	dryRun, err := ReadConfigBool(cmd, "dry-run")
	if err != nil {
		log.Printf("Error Processing - dry-run flag, error: %v", err)
		return flags, err
	}
	flags.DryRun = dryRun

	useTelemetry, err := ReadConfigBool(cmd, "use-telemetry")
	if err != nil {
		log.Printf("Error Processing - use-telemetry flag, error: %v", err)
		return flags, err
	}
	flags.UseTelemetry = useTelemetry

	silentMode, err := ReadConfigBool(cmd, "silent")
	if err != nil {
		log.Printf("Error Processing - use-telemetry flag, error: %v", err)
		return flags, err
	}

	flags.SilentMode = silentMode

	return flags, nil

}
