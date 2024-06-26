package pkg

import (
	"log"
	"os"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
)

// getChartFromRelease should get the manifests from the selected release.
func (image *Images) getChartFromRelease() ([]byte, error) {
	settings := cli.New()

	image.log.Debugf("fetching chart manifest for release '%s' from kube cluster", image.release)

	actionConfig := new(action.Configuration)

	if err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		image.log.Error("oops initialising helm client errored with", err)

		return nil, err
	}

	client := action.NewGet(actionConfig)

	image.log.Debugf("fetching manifests from revision '%d' of helm release '%s'", image.Revision, image.release)
	client.Version = image.Revision

	release, err := client.Run(image.release)
	if err != nil {
		return nil, err
	}

	image.log.Debugf("chart manifest for release '%s' was successfully retrieved from kube cluster", image.release)

	return []byte(release.Manifest), nil
}
