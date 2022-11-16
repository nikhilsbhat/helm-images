package pkg

import (
	"fmt"
	"strings"

	"github.com/nikhilsbhat/helm-images/pkg/k8s"
)

func (image *Images) filterImagesByRegistries(images []*k8s.Image) []*k8s.Image {
	if !image.UniqueImages && (len(image.Registries) == 0) {
		return images
	}

	var imagesFiltered []*k8s.Image
	if image.UniqueImages {
		image.log.Debug("limiting to unique images since '--unique/-u' is enabled")

		for _, img := range images {
			uniqueImages := getUniqEntries(img.Image)
			if len(uniqueImages) != 0 {
				img.Image = uniqueImages
				imagesFiltered = append(imagesFiltered, img)
			}
		}
	}

	var newImagesFiltered []*k8s.Image

	imagesToFilter := images
	if len(image.Registries) != 0 {
		if image.UniqueImages {
			imagesToFilter = imagesFiltered
		}
		image.log.Debug(fmt.Sprintf("filtering images by the selected registries '%s' since '-r,--registry' is enabled",
			strings.Join(image.Registries, ", ")))

		for _, img := range imagesToFilter {
			uniqueImages := filteredImages(img.Image, image.Registries)
			if len(uniqueImages) != 0 {
				img.Image = uniqueImages
				newImagesFiltered = append(newImagesFiltered, img)
			}
		}

		return newImagesFiltered
	}

	return imagesFiltered
}

func filteredImages(images []string, registries []string) []string {
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
