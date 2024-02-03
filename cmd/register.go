package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/nikhilsbhat/helm-images/pkg"
	imgErrors "github.com/nikhilsbhat/helm-images/pkg/errors"
	"github.com/nikhilsbhat/helm-images/version"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	images    = pkg.Images{}
	cliLogger *logrus.Logger
)

const (
	getArgumentCountLocal   = 2
	getArgumentCountRelease = 1
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
	command.commands = append(command.commands, getImagesCommand())
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

func getImagesCommand() *cobra.Command {
	imageCommand := &cobra.Command{
		Use:   "get [RELEASE] [CHART] [flags]",
		Short: "Fetches all images those are part of specified chart/release",
		Long:  "Lists all images those are part of specified chart/release and matches the pattern or part of specified registry.",
		Example: `  helm images get prometheus-standalone path/to/chart/prometheus-standalone -f ~/path/to/override-config.yaml
  helm images get prometheus-standalone --from-release --registry quay.io
  helm images get prometheus-standalone --from-release --registry quay.io --unique
  helm images get prometheus-standalone --from-release --registry quay.io --yaml
  helm images get oci://registry-1.docker.io/bitnamicharts/airflow --yaml
  helm images get kong-2.35.0.tgz --yaml`,
		Args:    validateAndSetArgs,
		PreRunE: setCLIClient,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			if (images.JSON && images.YAML && images.Table) || (images.JSON && images.YAML) ||
				(images.Table && images.YAML) || (images.Table && images.JSON) {
				return &imgErrors.MultipleFormatError{
					Message: "cannot render the output to multiple format, enable any of '--yaml --json --table' at a time",
				}
			}

			return images.GetImages()
		},
	}

	registerGetFlags(imageCommand)

	return imageCommand
}

func setCLIClient(_ *cobra.Command, _ []string) error {
	logger := logrus.New()
	logger.SetLevel(pkg.GetLoglevel(images.LogLevel))
	logger.WithField("helm-images", true)
	logger.SetFormatter(&logrus.JSONFormatter{})
	cliLogger = logger

	images.SetLogger(images.LogLevel)
	images.SetWriter(os.Stdout)

	return nil
}

func getRootCommand() *cobra.Command {
	rootCommand := &cobra.Command{
		Use:   "images [command]",
		Short: "Utility that helps in fetching images which are part of deployment",
		Long:  `Lists all images that would be part of helm deployment.`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Usage()
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

func versionConfig(_ *cobra.Command, _ []string) error {
	buildInfo, err := json.Marshal(version.GetBuildInfo())
	if err != nil {
		cliLogger.Fatalf("fetching version of helm-images failed with: %v", err)
	}

	writer := bufio.NewWriter(os.Stdout)
	versionInfo := fmt.Sprintf("%s \n", strings.Join([]string{"images version", string(buildInfo)}, ": "))

	if _, err = writer.WriteString(versionInfo); err != nil {
		cliLogger.Fatalln(err)
	}

	defer func(writer *bufio.Writer) {
		err = writer.Flush()
		if err != nil {
			cliLogger.Fatalln(err)
		}
	}(writer)

	return nil
}

//nolint:goerr113
func validateAndSetArgs(cmd *cobra.Command, args []string) error {
	logger := logrus.New()
	logger.SetLevel(pkg.GetLoglevel(images.LogLevel))
	logger.WithField("helm-images", true)
	logger.SetFormatter(&logrus.JSONFormatter{})
	cliLogger = logger

	minArgError := errors.New("[RELEASE] or [CHART] cannot be empty")
	oneOfThemError := errors.New("when '--from-release' is enabled, only [RELEASE] can be set and not both [RELEASE] [CHART]")
	defaultReleaseName := "sample"
	cmd.SilenceUsage = true

	if !images.FromRelease {
		switch len(args) {
		case getArgumentCountRelease:
			cliLogger.Debugf("looks like no release name specified, hence it would be set to '%s' by default", defaultReleaseName)

			images.SetRelease(defaultReleaseName)
			images.SetChart(args[0])
		case getArgumentCountLocal:
			images.SetRelease(args[0])
			images.SetChart(args[1])
		default:
			cliLogger.Fatal(minArgError)
		}

		return nil
	}

	if len(args) > getArgumentCountRelease {
		cliLogger.Fatal(oneOfThemError)
	}

	images.SetRelease(args[0])

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
