package pkg

import (
	"log"
	"os"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
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

	helmRelease, err := client.Run(image.release)
	if err != nil {
		return nil, err
	}

	image.log.Debugf("chart manifest for release '%s' was successfully retrieved from kube cluster", image.release)

	return []byte(helmRelease.Manifest), nil
}

func (image *Images) getChartsFromReleases() ([]*release.Release, error) {
	settings := cli.New()

	var namespace string

	if image.isAll() {
		image.log.Debug("no namespace specified, fetching all helm releases from the the cluster")
	} else {
		image.log.Debugf("retrieving charts from the namespace '%s'", image.namespace)
		namespace = image.namespace
	}

	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), namespace, os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		image.log.Error("oops initialising helm client errored with", err)

		return nil, err
	}

	client := action.NewList(actionConfig)

	return client.Run()
}

func (image *Images) isAll() bool {
	if image.namespace == "default" {
		return !(image.IsDefaultNamespace)
	}

	if len(image.namespace) != 0 {
		return false
	}

	return true
}
