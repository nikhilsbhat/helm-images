package pkg

import (
	"strings"

	"github.com/nikhilsbhat/helm-images/pkg/k8s"
)

func (image *Images) setOutput(images []*k8s.Image) any {
	images = image.FilterImagesByRegistriesNew(images)
	image.setImagePlatforms(images)

	var output any

	output = images

	if image.table {
		outputTable := make([][]string, 0)

		outputTable = append(outputTable, image.tableHeaders())
		for _, img := range images {
			outputTable = append(outputTable, image.tableRow(img))
		}

		output = outputTable
	}

	if !image.json && !image.yaml && !image.table && !image.csv {
		imagesNames := GetImagesFromKind(images)
		if image.UniqueImages {
			imagesNames = GetUniqEntries(imagesNames)
		}

		if image.Platform != "" {
			imagesNames = image.pullCommands(imagesNames)
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
		image.table = !image.all
		if image.all {
			image.log.Info("rendering results to 'table' format is not supported while fetching images from all releases, hence setting to default")
		}
	case "csv", "c":
		image.csv = true
	default:
		if len(image.OutputFormat) != 0 {
			image.log.Warnf("helm images does not support format '%s', switching to default", image.OutputFormat)
		}
	}
}

func (image *Images) setImagePlatforms(images []*k8s.Image) {
	if image.Platform == "" {
		return
	}

	for _, img := range images {
		img.Platform = image.Platform
	}
}

func (image *Images) tableHeaders() []string {
	if image.Platform == "" {
		return []string{"Name", "Kind", "Image"}
	}

	return []string{"Name", "Kind", "Platform", "Image"}
}

func (image *Images) tableRow(img *k8s.Image) []string {
	if image.Platform == "" {
		return []string{img.Name, img.Kind, strings.Join(img.Image, ", ")}
	}

	return []string{img.Name, img.Kind, img.Platform, strings.Join(img.Image, ", ")}
}

func (image *Images) pullCommands(imagesNames []string) []string {
	pullCommands := make([]string, 0, len(imagesNames))

	for _, imageName := range imagesNames {
		pullCommands = append(pullCommands, "docker pull --platform "+image.Platform+" "+imageName)
	}

	return pullCommands
}
