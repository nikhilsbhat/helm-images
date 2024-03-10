package pkg

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"strings"

	imageErr "github.com/nikhilsbhat/helm-images/pkg/errors"
	"github.com/sirupsen/logrus"
)

// getChartFromTemplate should get the manifests by rendering the helm template.
func (image *Images) getChartFromTemplate() ([]byte, error) {
	flags := make([]string, 0)

	for _, value := range image.Values {
		flags = append(flags, "--set", value)
	}

	for _, stringValue := range image.StringValues {
		flags = append(flags, "--set-string", stringValue)
	}

	for _, showOnly := range image.ShowOnly {
		flags = append(flags, "--show-only", showOnly)
	}

	for _, fileValue := range image.FileValues {
		flags = append(flags, "--set-file", fileValue)
	}

	for _, valueFile := range image.ValueFiles {
		flags = append(flags, "--values", valueFile)
	}

	if strings.ToLower(image.LogLevel) == logrus.DebugLevel.String() {
		flags = append(flags, "--debug")
	}

	if image.SkipTests {
		flags = append(flags, "--skip-tests")
	}

	if image.SkipCRDS {
		flags = append(flags, "--skip-crds")
	}

	if image.Validate {
		flags = append(flags, "--validate")
	}

	if len(image.Version) != 0 {
		flags = append(flags, "--version", image.Version)
	}

	args := []string{"template", image.release, image.chart}
	args = append(args, flags...)

	image.log.Debugf("rendering helm chart with following commands/flags '%s'", strings.Join(args, ", "))

	helmBin := os.Getenv("HELM_BIN")
	if helmBin == "" {
		return nil, &imageErr.ImageError{Message: "environment variable 'HELM_BIN' is not set"}
	}

	cmd := exec.Command(helmBin, args...)
	image.log.Debugf("running following command to render the helm template: %s", cmd.String())
	output, err := cmd.Output()

	var exitErr *exec.ExitError

	if errors.As(err, &exitErr) {
		image.log.Errorf("rendering template for release: '%s' errored with %v", image.release, err)

		return nil, fmt.Errorf("%w: %s", exitErr, exitErr.Stderr)
	}

	var pathErr *fs.PathError

	if errors.As(err, &pathErr) {
		image.log.Error("locating helm cli errored with", err)

		return nil, fmt.Errorf("%w: %s", pathErr, pathErr.Path)
	}

	return output, nil
}
