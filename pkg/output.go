package pkg

import (
	"strings"

	"github.com/nikhilsbhat/helm-images/pkg/k8s"
)

func (image *Images) setOutput(images []*k8s.Image) interface{} {
	images = image.FilterImagesByRegistriesNew(images)

	var output interface{}
	output = images

	if image.table {
		outputTable := make([][]string, 0)

		outputTable = append(outputTable, []string{"Name", "Kind", "Image"})
		for _, img := range images {
			outputTable = append(outputTable, []string{img.Name, img.Kind, strings.Join(img.Image, ", ")})
		}

		output = outputTable
	}

	if !image.json && !image.yaml && !image.table && !image.csv {
		imagesNames := GetImagesFromKind(images)
		if image.UniqueImages {
			imagesNames = GetUniqEntries(imagesNames)
		}

		output = strings.Join(imagesNames, "\n")
	}

	return output
}

func (image *Images) SetOutputFormats() {
	switch strings.ToLower(image.OutputFormat) {
	case "yaml", "y":
		image.yaml = true
	case "json", "j":
		image.json = true
	case "table", "t":
		image.table = true
	case "csv", "c":
		image.csv = true
	default:
		if len(image.OutputFormat) != 0 {
			image.log.Warnf("helm images does not support format '%s', switching to default", image.OutputFormat)
		}
	}
}
