package pkg

import (
	"strings"

	"github.com/nikhilsbhat/helm-images/pkg/k8s"
)

// FilterImagesByRegistries filters images those are part of registries list passed.
func (image *Images) FilterImagesByRegistries(images []*k8s.Image) []*k8s.Image {
	if !image.UniqueImages && (len(image.Registries) == 0) {
		return images
	}

	var imagesFiltered []*k8s.Image

	if image.UniqueImages {
		image.log.Debug("limiting to unique images since '--unique/-u' is enabled")

		for _, img := range images {
			uniqueImages := GetUniqEntries(img.Image)
			if len(uniqueImages) != 0 {
				img.Image = uniqueImages
				imagesFiltered = append(imagesFiltered, img)
			}
		}
	}

	if len(image.Registries) != 0 {
		var newImagesFiltered []*k8s.Image

		imagesToFilter := images

		if image.UniqueImages {
			imagesToFilter = imagesFiltered
		}

		image.log.Debugf("filtering images by the selected registries '%s' since '-r,--registry' is enabled",
			strings.Join(image.Registries, ", "))

		for _, img := range imagesToFilter {
			uniqueImages := FilteredImages(img.Image, image.Registries)
			if len(uniqueImages) != 0 {
				img.Image = uniqueImages
				newImagesFiltered = append(newImagesFiltered, img)
			}
		}

		return newImagesFiltered
	}

	return imagesFiltered
}

func (image *Images) FilterImagesByRegistriesNew(images []*k8s.Image) []*k8s.Image {
	if !image.UniqueImages && len(image.Registries) == 0 {
		return images
	}

	image.log.Debug("starting image filtering")

	imagesFiltered := images

	if image.UniqueImages {
		image.log.Debug("limiting to unique images since '--unique/-u' is enabled")

		imagesFiltered = filterUniqueImages(images)
	}

	if len(image.Registries) != 0 {
		image.log.Debugf("filtering images by the selected registries '%s' since '-r,--registry' is enabled",
			strings.Join(image.Registries, ", "))

		imagesFiltered = filterByRegistries(imagesFiltered, image.Registries)
	}

	return imagesFiltered
}

func filterUniqueImages(images []*k8s.Image) []*k8s.Image {
	var imagesFiltered []*k8s.Image

	for _, img := range images {
		uniqueImages := GetUniqEntries(img.Image)
		if len(uniqueImages) != 0 {
			img.Image = uniqueImages
			imagesFiltered = append(imagesFiltered, img)
		}
	}

	return imagesFiltered
}

func filterByRegistries(images []*k8s.Image, registries []string) []*k8s.Image {
	var imagesFiltered []*k8s.Image

	for _, img := range images {
		filteredImages := filterImagesByRegistries(img.Image, registries)
		if len(filteredImages) != 0 {
			img.Image = filteredImages
			imagesFiltered = append(imagesFiltered, img)
		}
	}

	return imagesFiltered
}

func filterImagesByRegistries(images, registries []string) []string {
	var imagesFiltered []string

	for _, image := range images {
		for _, registry := range registries {
			if strings.HasPrefix(image, registry) {
				imagesFiltered = append(imagesFiltered, image)
			}
		}
	}

	return imagesFiltered
}

func FilteredImages(images, registries []string) []string {
	var imagesFiltered []string

	for _, registry := range registries {
		for _, image := range images {
			if strings.HasPrefix(image, registry) {
				imagesFiltered = append(imagesFiltered, image)
			}
		}
	}

	return imagesFiltered
}
