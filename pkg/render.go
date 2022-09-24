package pkg

import (
	"encoding/json"
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/nikhilsbhat/helm-images/pkg/k8s"
)

func (image *Images) render(images []*k8s.Image) error {
	if image.JSON {
		kindJSON, err := json.MarshalIndent(images, " ", " ")
		if err != nil {
			return err
		}
		fmt.Printf("%s", string(kindJSON))

		return nil
	}

	if image.YAML {
		kindYAML, err := yaml.Marshal(images)
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", "---")
		fmt.Printf("%s", string(kindYAML))

		return nil
	}

	imagesFromKind := getImagesFromKind(images)
	filteredImages := image.filterImagesByRegistries(imagesFromKind)

	if image.UniqueImages {
		filteredImages = getUniqEntries(filteredImages)
	}

	for _, img := range filteredImages {
		fmt.Printf("%v\n", img)
	}

	return nil
}
