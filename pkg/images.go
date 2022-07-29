package pkg

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/nikhilsbhat/helm-images/pkg/k8s"

	"github.com/spf13/cobra"
	"github.com/thoas/go-funk"
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
	JSON         bool
	YAML         bool
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

	images := make([]*k8s.Image, 0)
	kubeKindTemplates := image.getTemplates(chart)
	for _, kubeKindTemplate := range kubeKindTemplates {
		currentKind, err := k8s.NewKind().Get(kubeKindTemplate)
		if err != nil {
			return err
		}

		if !funk.Contains(image.Kind, currentKind) {
			continue
		}

		switch currentKind {
		case k8s.KindDeployment:
			deployImages, err := k8s.NewDeployment().Get(kubeKindTemplate)
			if err != nil {
				return err
			}
			images = append(images, deployImages)
		case k8s.KindStatefulSet:
			stsImages, err := k8s.NewStatefulSet().Get(kubeKindTemplate)
			if err != nil {
				return err
			}
			images = append(images, stsImages)
		case k8s.KindDaemonSet:
			daemonImages, err := k8s.NewDaemonSet().Get(kubeKindTemplate)
			if err != nil {
				return err
			}
			images = append(images, daemonImages)
		case k8s.KindReplicaSet:
			replicaSets, err := k8s.NewReplicaSets().Get(kubeKindTemplate)
			if err != nil {
				return err
			}
			images = append(images, replicaSets)
		case k8s.KindCronJob:
			cronJob, err := k8s.NewCronjob().Get(kubeKindTemplate)
			if err != nil {
				return err
			}
			images = append(images, cronJob)
		case k8s.KindJob:
			job, err := k8s.NewJob().Get(kubeKindTemplate)
			if err != nil {
				return err
			}
			images = append(images, job)
		default:
			log.Printf("kind %v is not supported at the moment", currentKind)
		}
	}
	return image.render(images)
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

func (image *Images) filterImagesByRegistries(images []string) []string {
	if len(image.Registries) == 0 {
		return images
	}

	var filteredImages []string
	for _, registry := range image.Registries {
		for _, img := range images {
			if strings.HasPrefix(img, registry) {
				filteredImages = append(filteredImages, img)
			}
		}
	}
	return filteredImages
}

func getImagesFromKind(kinds []*k8s.Image) (images []string) {
	for _, knd := range kinds {
		images = append(images, knd.Image...)
	}
	return
}
