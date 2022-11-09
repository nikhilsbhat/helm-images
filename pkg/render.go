package pkg

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cheynewallace/tabby"

	"github.com/ghodss/yaml"
	"github.com/nikhilsbhat/helm-images/pkg/k8s"
)

func (image *Images) render(images []*k8s.Image) error {
	imagesFiltered := image.filterImagesByRegistries(images)

	if image.JSON {
		kindJSON, err := json.MarshalIndent(imagesFiltered, " ", " ")
		if err != nil {
			return err
		}
		fmt.Printf("%s", string(kindJSON))

		return nil
	}

	if image.YAML {
		kindYAML, err := yaml.Marshal(imagesFiltered)
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", "---")
		fmt.Printf("%s", string(kindYAML))

		return nil
	}

	if image.Table {
		table := tabby.New()
		table.AddHeader("Name", "Kind", "Image")
		for _, img := range imagesFiltered {
			table.AddLine(img.Name, img.Kind, strings.Join(img.Image, ", "))
		}
		table.Print()

		return nil
	}

	var imgs []string
	for _, img := range imagesFiltered {
		imgs = append(imgs, img.Image...)
	}

	if image.UniqueImages {
		imgs = getUniqEntries(imgs)
	}

	for _, img := range imgs {
		fmt.Printf("%v\n", img)
	}

	return nil
}
