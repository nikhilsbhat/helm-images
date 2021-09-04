package pkg

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
)

const (
	// ImageRegex is the default regex, that is used to split one big helm template to multiple templates.
	// Splitting templates eases the task of  identifying Kubernetes objects.
	ImageRegex = `---\n# Source:\s.*.`
)

// Images represents GetImages.
type Images struct {
	Registries   []string
	Kind         []string
	Values       []string
	StringValues []string
	FileValues   []string
	ImageRegex   string
	ValueFiles   ValueFiles
	UniqueImages bool
	release      string
	chart        string
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

// GetImages fetches all available images from the specified chart.
// Also filters identified images, to get just unique ones.
func (image *Images) GetImages(cmd *cobra.Command, args []string) error {
	cmd.SilenceUsage = true

	image.release = args[0]
	image.chart = args[1]

	chart, err := image.getChartTemplate()
	if err != nil {
		return err
	}

	selectedKinds := make([]map[string]interface{}, 0)
	images := make([]string, 0)
	kubeKindTemplates := image.getTemplates(chart)
	for _, kubeKindTemplate := range kubeKindTemplates {
		var kindYaml map[string]interface{}
		if err := yaml.Unmarshal([]byte(kubeKindTemplate), &kindYaml); err != nil {
			return err
		}
		if len(image.Kind) != 0 {
			if find(image.Kind, kindYaml["kind"].(string)) {
				selectedKinds = append(selectedKinds, kindYaml)
			}
		} else {
			selectedKinds = append(selectedKinds, kindYaml)
		}
	}

	for _, selectedKind := range selectedKinds {
		if foundImage, ok := findKey(selectedKind, "image"); ok {
			images = append(images, foundImage.(string))
		}
	}

	filteredImages := image.filterImages(images)

	if image.UniqueImages {
		filteredImages = getUniqueSlice(filteredImages)
	}

	for _, img := range filteredImages {
		fmt.Printf("%v\n", img)
	}
	return nil
}

func (image *Images) getChartTemplate() ([]byte, error) {
	flags := make([]string, 0)
	for _, value := range image.Values {
		flags = append(flags, "--set", value)
	}
	for _, stringValue := range image.StringValues {
		flags = append(flags, "--set-string", stringValue)
	}
	for _, fileValue := range image.FileValues {
		flags = append(flags, "--set-file", fileValue)
	}
	for _, valueFile := range image.ValueFiles {
		flags = append(flags, "--values", valueFile)
	}

	args := []string{"template", image.release, image.chart}
	args = append(args, flags...)

	cmd := exec.Command(os.Getenv("HELM_BIN"), args...) //nolint:gosec
	output, err := cmd.Output()
	if exitError, ok := err.(*exec.ExitError); ok {
		return nil, fmt.Errorf("%s: %s", exitError.Error(), string(exitError.Stderr))
	}
	if pathError, ok := err.(*fs.PathError); ok {
		return nil, fmt.Errorf("%s: %s", pathError.Error(), pathError.Path)
	}
	return output, nil
}

func (image *Images) getTemplates(template []byte) []string {
	temp := regexp.MustCompile(image.ImageRegex)
	kinds := temp.Split(string(template), -1)
	// Removing empty string at the beginning as splitting string always adds it in front.
	kinds = kinds[1:]
	return kinds
}

func (image *Images) filterImages(images []string) (filteredImages []string) {
	if len(image.Registries) == 0 {
		return images
	}
	for _, registry := range image.Registries {
		for _, img := range images {
			if strings.HasPrefix(img, registry) {
				filteredImages = append(filteredImages, img)
			}
		}
	}
	return
}
