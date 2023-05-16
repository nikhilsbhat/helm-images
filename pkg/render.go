package pkg

import (
	"bufio"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cheynewallace/tabby"
	"github.com/ghodss/yaml"
	"github.com/nikhilsbhat/helm-images/pkg/k8s"
)

func (image *Images) render(images []*k8s.Image) error {
	imagesFiltered := image.FilterImagesByRegistries(images)

	if image.JSON {
		if err := image.ToJSON(imagesFiltered); err != nil {
			return err
		}

		return nil
	}

	if image.YAML {
		if err := image.ToYAML(imagesFiltered); err != nil {
			return err
		}

		return nil
	}

	if image.Table {
		image.toTABLE(imagesFiltered)

		return nil
	}

	image.log.Debug("no formart was specified for rendering images, defaulting to list")

	imags := GetImagesFromKind(imagesFiltered)

	if image.UniqueImages {
		imags = GetUniqEntries(imags)
	}

	if _, err := image.writer.Write([]byte(fmt.Sprintf("%s\n", strings.Join(imags, "\n")))); err != nil {
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

	table := tabby.New()
	table.AddHeader("Name", "Kind", "Image")

	for _, img := range imagesFiltered {
		table.AddLine(img.Name, img.Kind, strings.Join(img.Image, ", "))
	}

	table.Print()
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
