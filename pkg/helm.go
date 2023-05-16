package pkg

import (
	"fmt"
	"log"
	"os"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
)

// GetImagesFromRelease should get the manifest from the selected release.
func (image *Images) GetImagesFromRelease() ([]byte, error) {
	settings := cli.New()

	image.log.Debug(fmt.Sprintf("fetching chart manifest for release '%s' from kube cluster", image.release))

	actionConfig := new(action.Configuration)

	if err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		image.log.Error("oops initialising helm client errored with", err)

		return nil, err
	}

	client := action.NewGet(actionConfig)

	release, err := client.Run(image.release)
	if err != nil {
		return nil, err
	}

	image.log.Debug(fmt.Sprintf("chart manifest for release '%s' was successfully retrieved from kube cluster", image.release))

	return []byte(release.Manifest), nil
}
