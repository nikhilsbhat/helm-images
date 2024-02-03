package pkg

import (
	"strings"

	"github.com/nikhilsbhat/helm-images/pkg/k8s"
)

func (image *Images) setOutput(images []*k8s.Image) interface{} {
	images = image.FilterImagesByRegistries(images)

	var output interface{}
	output = images

	if image.Table {
		outputTable := make([][]string, 0)

		outputTable = append(outputTable, []string{"Name", "Kind", "Image"})
		for _, img := range images {
			outputTable = append(outputTable, []string{img.Name, img.Kind, strings.Join(img.Image, ", ")})
		}

		output = outputTable
	}

	if !image.JSON && !image.YAML && !image.Table {
		imagesNames := GetImagesFromKind(images)
		if image.UniqueImages {
			imagesNames = GetUniqEntries(imagesNames)
		}

		output = strings.Join(imagesNames, "\n")
	}

	return output
}
