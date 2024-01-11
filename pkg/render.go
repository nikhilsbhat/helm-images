package pkg

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/nikhilsbhat/helm-images/pkg/k8s"
	"github.com/olekukonko/tablewriter"
)

func (image *Images) render(k8sImages []*k8s.Image) error {
	imagesFiltered := image.FilterImagesByRegistries(k8sImages)

	if image.JSON {
		return image.ToJSON(imagesFiltered)
	}

	if image.YAML {
		return image.ToYAML(imagesFiltered)
	}

	if image.Table {
		image.toTABLE(imagesFiltered)

		return nil
	}

	image.log.Debug("no format was specified for rendering images, defaulting to list")

	images := GetImagesFromKind(imagesFiltered)

	if image.UniqueImages {
		images = GetUniqEntries(images)
	}

	if _, err := image.writer.Write([]byte(fmt.Sprintf("%s\n", strings.Join(images, "\n")))); err != nil {
		image.log.Fatalln(err)
	}

	defer func(writer *bufio.Writer) {
		err := writer.Flush()
		if err != nil {
			image.log.Fatalln(err)
		}
	}(image.writer)

	return nil
}

func (image *Images) toTABLE(imagesFiltered []*k8s.Image) {
	image.log.Debug("rendering the images in table format since --table is enabled")

	table := tablewriter.NewWriter(os.Stdout)

	table.SetHeader([]string{"Name", "Kind", "Image"})

	if !image.NoColor {
		table.SetHeaderColor(tablewriter.Colors{tablewriter.Bold}, tablewriter.Colors{tablewriter.Bold}, tablewriter.Colors{tablewriter.Bold})
	}

	table.SetAlignment(tablewriter.ALIGN_CENTER) //nolint:nosnakecase
	table.SetAutoWrapText(true)
	table.SetAutoMergeCells(true)
	table.SetRowLine(true)

	for _, img := range imagesFiltered {
		table.Append([]string{img.Name, img.Kind, strings.Join(img.Image, ", ")})
	}

	table.Render()
}

func (image *Images) ToYAML(imagesFiltered []*k8s.Image) error {
	image.log.Debug("rendering the images in yaml format since --yaml is enabled")

	kindYAML, err := yaml.Marshal(imagesFiltered)
	if err != nil {
		return err
	}

	yamlString := strings.Join([]string{"---", string(kindYAML)}, "\n")

	if _, err = image.writer.Write([]byte(yamlString)); err != nil {
		image.log.Fatalln(err)
	}

	defer func(writer *bufio.Writer) {
		err = writer.Flush()
		if err != nil {
			image.log.Fatalln(err)
		}
	}(image.writer)

	return nil
}

func (image *Images) ToJSON(imagesFiltered []*k8s.Image) error {
	image.log.Debug("rendering the images in json format since --json is enabled")

	kindJSON, err := json.MarshalIndent(imagesFiltered, " ", " ")
	if err != nil {
		return err
	}

	if _, err = image.writer.Write(kindJSON); err != nil {
		image.log.Fatalln(err)
	}

	defer func(writer *bufio.Writer) {
		err = writer.Flush()
		if err != nil {
			image.log.Fatalln(err)
		}
	}(image.writer)

	return nil
}
