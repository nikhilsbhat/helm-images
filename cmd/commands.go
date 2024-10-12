package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/nikhilsbhat/helm-images/pkg"
	"github.com/nikhilsbhat/helm-images/version"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func getRootCommand() *cobra.Command {
	rootCommand := &cobra.Command{
		Use:   "images [command]",
		Short: "Utility that helps in fetching images which are part of deployment",
		Long:  `Lists all images that would be part of helm deployment.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Usage()
		},
	}
	rootCommand.SetUsageTemplate(getUsageTemplate())

	return rootCommand
}

func getImagesCommand() *cobra.Command {
	imageCommand := &cobra.Command{
		Use:   "get [RELEASE] [CHART] [flags]",
		Short: "Fetches all images those are part of specified chart/release",
		Long:  "Lists all images those are part of specified chart/release and matches the pattern or part of specified registry.",
		Example: `  helm images get prometheus-standalone path/to/chart/prometheus-standalone -f ~/path/to/override-config.yaml
  helm images get prometheus-standalone --from-release --registry quay.io -o table
  helm images get prometheus-standalone --from-release --registry quay.io --unique
  helm images get prometheus-standalone --from-release --registry quay.io -o yaml
  helm images get oci://registry-1.docker.io/bitnamicharts/airflow -o yaml
  helm images get kong-2.35.0.tgz -o json
  helm template example/chart/sample | helm images get --raw -
  helm template example/chart/sample | helm images get --raw - -o yaml`,
		Args:    validateAndSetArgs,
		PreRunE: setCLIClient,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.SilenceUsage = true

			if images.Raw {
				stdIn := cmd.InOrStdin()
				imagesRaw, err := io.ReadAll(stdIn)
				if err != nil {
					return err
				}

				images.SetRaw(imagesRaw)
			}

			return images.GetImages()
		},
	}

	registerCommonFlags(imageCommand)
	registerGetFlags(imageCommand)

	imageCommand.MarkFlagsMutuallyExclusive("raw", "from-release")

	return imageCommand
}

func getAllImagesCommand() *cobra.Command {
	allImageCommand := &cobra.Command{
		Use:   "all [flags]",
		Short: "Fetches all images from all release",
		Long:  "Lists all images part of all release present in the cluster with matching pattern or part of specified registry.",
		Example: `  helm images all -o yaml
  helm images all -n monitoring -o yaml
  helm images all -n monitoring -o yaml --unique
  helm images all --skip-release traefik=kube-system --registry quay.io -o json`,
		Args:    cobra.NoArgs,
		PreRunE: setCLIClient,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.SilenceUsage = true

			images.SetNamespace(os.Getenv("HELM_NAMESPACE"))

			if err := images.SetReleasesToSkips(); err != nil {
				return err
			}

			return images.GetAllImages()
		},
	}

	registerCommonFlags(allImageCommand)
	registerGetAllFlags(allImageCommand)

	return allImageCommand
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

func setCLIClient(cmd *cobra.Command, _ []string) error {
	logger := logrus.New()
	logger.SetLevel(pkg.GetLoglevel(images.LogLevel))
	logger.WithField("helm-images", true)
	logger.SetFormatter(&logrus.JSONFormatter{})
	cliLogger = logger

	images.SetLogger(images.LogLevel)

	if cmd.Use == "all [flags]" {
		images.SetAll(true)
	}

	images.SetOutputFormats()

	images.SetRenderer()

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

	if images.Raw {
		return nil
	}

	if images.Revision != 0 && !images.FromRelease {
		cliLogger.Fatalf("the '--revision' flag can only be used when retrieving images from a release, i.e., when the '--from-release' flag is set")
	}

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

	if len(args) == 0 {
		cliLogger.Fatal("[RELEASE] name missing")
	}

	images.SetRelease(args[0])

	return nil
}
