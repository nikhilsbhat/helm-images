package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/nikhilsbhat/helm-images/pkg"
	"github.com/nikhilsbhat/helm-images/version"
	"github.com/spf13/cobra"
)

var (
	images = pkg.Images{}
)

type imagesCommands struct {
	commands []*cobra.Command
}

// SetImagesCommands helps in gathering all the subcommands so that it can be used while registering it with main command.
func SetImagesCommands() *cobra.Command {
	return getImagesCommands()
}

// Add an entry in below function to register new command.
func getImagesCommands() *cobra.Command {
	command := new(imagesCommands)
	command.commands = append(command.commands, getImagesCommnd())
	command.commands = append(command.commands, getVersionCommand())
	return command.prepareCommands()
}

func (c *imagesCommands) prepareCommands() *cobra.Command {
	rootCmd := getRootCommand()
	for _, cmnd := range c.commands {
		rootCmd.AddCommand(cmnd)
	}
	registerFlags(rootCmd)
	return rootCmd
}

func getImagesCommnd() *cobra.Command {
	imageCommand := &cobra.Command{
		Use:   "get [RELEASE] [CHART] [flags]",
		Short: "Fetches all images part of deployment",
		Long:  "Lists all images that matches the pattern or part of specified registry.",
		Args:  minimumArgError,
		RunE:  images.GetImages,
	}
	registerGetFlags(imageCommand)
	return imageCommand
}

func getRootCommand() *cobra.Command {
	rootCommand := &cobra.Command{
		Use:   "images [command]",
		Short: "Utility that helps in fetching all images",
		Long:  `Lists all images that would be part of helm deployment would be listed.`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Usage(); err != nil {
				return err
			}
			return nil
		},
	}
	rootCommand.SetUsageTemplate(getUsageTemplate())
	return rootCommand
}

func getVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version [flags]",
		Short: "Command to fetch the version of helm-images installed",
		Long:  `This will help user to find what version of helm-images plugin he/she installed in her machine.`,
		RunE:  versionConfig,
	}
}

func versionConfig(cmd *cobra.Command, args []string) error {
	buildInfo, err := json.Marshal(version.GetBuildInfo())
	if err != nil {
		log.Fatalf("fetching version of helm-images failed with: %v", err)
		os.Exit(1)
	}
	fmt.Println("images version:", string(buildInfo))
	return nil
}

func minimumArgError(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true
	if len(args) != pkg.GetArgumentCount {
		log.Println("[RELEASE] or [CHART] cannot be empty")
		return fmt.Errorf("[RELEASE] or [CHART] cannot be empty")
	}
	return nil
}

func getUsageTemplate() string {
	return `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if gt (len .Aliases) 0}}{{printf "\n" }}
Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}{{printf "\n" }}
Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{printf "\n"}}
Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}{{printf "\n"}}
Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}{{printf "\n"}}
Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}{{printf "\n"}}
Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}
{{if .HasAvailableSubCommands}}{{printf "\n"}}
Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
{{printf "\n"}}`
}
