package pkg

import (
	"fmt"
	"strings"

	"github.com/nikhilsbhat/helm-images/pkg/errors"
	"github.com/nikhilsbhat/helm-images/pkg/k8s"
	"github.com/thoas/go-funk"
)

type skipReleaseInfo struct {
	name      string
	namespace string
}

func (image *Images) GetAllImages() error {
	releases, err := image.getChartsFromReleases()
	if err != nil {
		return err
	}

	releases = releasesToSkip(image.releasesToSkip).filterRelease(releases)

	imagesFromAllRelease := make([]k8s.Images, 0)

	for _, release := range releases {
		image.log.Debugf("fetching the images from release '%s' of namespace '%s'", release.Name, release.Namespace)

		images := make([]*k8s.Image, 0)
		kubeKindTemplates, err := image.GetTemplates([]byte(release.Manifest))
		if err != nil {
			return err
		}

		skips := image.GetResourcesToSkip()

		for _, kubeKindTemplate := range kubeKindTemplates {
			currentManifestName, err := k8s.NewName().Get(kubeKindTemplate, image.log)
			if err != nil {
				return err
			}

			currentKind, err := k8s.NewKind().Get(kubeKindTemplate, image.log)
			if err != nil {
				return err
			}

			if !funk.Contains(image.Kind, currentKind) {
				image.log.Debugf("either helm-images plugin does not support kind '%s' "+
					"at the moment or manifest might not have images to filter", currentKind)

				continue
			}

			shouldSkip := false

			for _, skip := range skips {
				if skip.Name == strings.ToLower(currentManifestName) && skip.Kind == strings.ToLower(currentKind) {
					image.log.Debugf("Skipping '%s' bearing name '%s' since it is set to skip.", currentKind, currentManifestName)

					shouldSkip = true

					break
				}
			}

			if shouldSkip {
				continue
			}

			image.log.Debugf("fetching images from '%s' of kind '%s'", currentKind, currentManifestName)

			imagesFound, err := image.GetImage(currentKind, kubeKindTemplate)
			if err != nil {
				return err
			}

			images = append(images, imagesFound...)
		}

		if len(images) == 0 {
			image.log.Infof("the release '%s' of namespace '%s' does not have any images", release.Name, release.Namespace)

			continue
		}

		output := image.setOutput(images)

		imagesFromAllRelease = append(imagesFromAllRelease, k8s.Images{ImagesFromRelease: output, NameSpace: release.Namespace})
	}

	return image.renderer.Render(imagesFromAllRelease)
}

func (image *Images) SetReleasesToSkips() error {
	const resourceLength = 2

	releasesToBeSkipped := make([]skipReleaseInfo, len(image.SkipReleases))

	for index, skipRelease := range image.SkipReleases {
		parsedRelease := strings.SplitN(skipRelease, "=", resourceLength)
		if len(parsedRelease) != resourceLength {
			return &errors.ImageError{Message: fmt.Sprintf("unable to parse release skip '%s'", skipRelease)}
		}

		releasesToBeSkipped[index] = skipReleaseInfo{name: parsedRelease[0], namespace: parsedRelease[1]}
	}

	image.releasesToSkip = releasesToBeSkipped

	return nil
}
