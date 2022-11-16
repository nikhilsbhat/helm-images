package pkg

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"

	"github.com/cheynewallace/tabby"
	"github.com/ghodss/yaml"
	"github.com/nikhilsbhat/helm-images/pkg/k8s"
)

func (image *Images) render(images []*k8s.Image) error {
	imagesFiltered := image.filterImagesByRegistries(images)

	writer := bufio.NewWriter(os.Stdout)
	if image.JSON {
		if err := image.toJSON(writer, imagesFiltered); err != nil {
			return err
		}

		return nil
	}

	if image.YAML {
		if err := image.toYAML(writer, imagesFiltered); err != nil {
			return err
		}

		return nil
	}

	if image.Table {
		image.toTABLE(imagesFiltered)

		return nil
	}

	image.log.Debug("no formart was specified for rendering images, defaulting to list")
	imgs := getImagesFromKind(imagesFiltered)

	if image.UniqueImages {
		imgs = getUniqEntries(imgs)
	}

	_, err := writer.Write([]byte(strings.Join(imgs, "\n")))
	if err != nil {
		image.log.Fatalln(err)
	}

	defer func(writer *bufio.Writer) {
		err = writer.Flush()
		if err != nil {
			image.log.Fatalln(err)
		}
	}(writer)

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

func (image *Images) toYAML(writer *bufio.Writer, imagesFiltered []*k8s.Image) error {
	image.log.Debug("rendering the images in yaml format since --yaml is enabled")
	kindYAML, err := yaml.Marshal(imagesFiltered)
	if err != nil {
		return err
	}

	yamlString := strings.Join([]string{"---", string(kindYAML)}, "\n")

	_, err = writer.Write([]byte(yamlString))
	if err != nil {
		image.log.Fatalln(err)
	}

	defer func(writer *bufio.Writer) {
		err = writer.Flush()
		if err != nil {
			image.log.Fatalln(err)
		}
	}(writer)

	return nil
}

func (image *Images) toJSON(writer *bufio.Writer, imagesFiltered []*k8s.Image) error {
	image.log.Debug("rendering the images in json format since --json is enabled")
	kindJSON, err := json.MarshalIndent(imagesFiltered, " ", " ")
	if err != nil {
		return err
	}

	_, err = writer.Write(kindJSON)
	if err != nil {
		image.log.Fatalln(err)
	}

	defer func(writer *bufio.Writer) {
		err = writer.Flush()
		if err != nil {
			image.log.Fatalln(err)
		}
	}(writer)

	return nil
}
