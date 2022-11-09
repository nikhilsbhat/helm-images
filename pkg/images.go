package pkg

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/nikhilsbhat/helm-images/pkg/k8s"

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
	FromRelease  bool
	UniqueImages bool
	JSON         bool
	YAML         bool
	Table        bool
	release      string
	chart        string
	log          *logrus.Logger
}

func (image *Images) SetRelease(release string) {
	image.release = release
}

func (image *Images) SetChart(chart string) {
	image.chart = chart
}

// GetImages fetches all available images from the specified chart.
// Also filters identified images, to get just unique ones.
func (image *Images) GetImages() error {
	image.log.Debug(
		fmt.Sprintf("got all required values to fetch the images from chart/release '%s' proceeding furter to fetch the same", image.release),
	)

	chart, err := image.getChartManifests()
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
			image.log.Debug(fmt.Sprintf("either helm-images plugin does not support kind '%s' "+
				"at the moment or manifest might not have images to filter", currentKind))

			continue
		}

		image.log.Debug(fmt.Sprintf("fetching images from kind '%s'", currentKind))

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
		case k8s.KindPod:
			pods, err := k8s.NewPod().Get(kubeKindTemplate)
			if err != nil {
				return err
			}
			images = append(images, pods)
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
			image.log.Debug(fmt.Sprintf("kind '%s' is not supported at the moment", currentKind))
		}
	}

	return image.render(images)
}

func (image *Images) getChartManifests() ([]byte, error) {
	if image.FromRelease {
		image.log.Debug(fmt.Sprintf("from-release is selected, hence fetching manifests for '%s' from helm release", image.release))

		return image.GetImagesFromRelease()
	}

	image.log.Debug(fmt.Sprintf("fetching manifests for '%s' by rendering helm template locally", image.release))

	return image.getChartTemplate()
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
		image.log.Error(fmt.Sprintf("rendering template for release: '%s' errored with ", image.release), err)

		return nil, fmt.Errorf("%w: %s", exitError, exitError.Stderr)
	}
	if pathError, ok := err.(*fs.PathError); ok {
		image.log.Error("locating helm cli errored with", err)

		return nil, fmt.Errorf("%w: %s", pathError, pathError.Path)
	}

	return output, nil
}

func (image *Images) getTemplates(template []byte) []string {
	image.log.Debug(fmt.Sprintf("splitting helm manifests with regex pattern: '%s'", image.ImageRegex))
	temp := regexp.MustCompile(image.ImageRegex)
	kinds := temp.Split(string(template), -1)
	// Removing empty string at the beginning as splitting string always adds it in front.
	kinds = kinds[1:]

	return kinds
}

func (image *Images) filterImagesByRegistries(images []*k8s.Image) []*k8s.Image {
	if !image.UniqueImages && (len(image.Registries) == 0) {
		return images
	}

	var imagesFiltered []*k8s.Image

	if image.UniqueImages {
		for _, img := range images {
			uniqueImages := getUniqEntries(img.Image)
			if len(uniqueImages) != 0 {
				img.Image = uniqueImages
				imagesFiltered = append(imagesFiltered, img)
			}
		}
		return imagesFiltered
	}

	var newImagesFiltered []*k8s.Image
	if len(image.Registries) != 0 {
		for _, img := range images {
			uniqueImages := filteredImages(img.Image, image.Registries)
			if len(uniqueImages) != 0 {
				img.Image = uniqueImages
				newImagesFiltered = append(newImagesFiltered, img)
			}
		}
	}

	return newImagesFiltered
}

func filteredImages(images []string, registries []string) (imagesFiltered []string) {
	for _, registry := range registries {
		for _, image := range images {
			if strings.HasPrefix(image, registry) {
				imagesFiltered = append(imagesFiltered, image)
			}
		}
	}

	return
}

func getImagesFromKind(kinds []*k8s.Image) []string {
	var images []string
	for _, knd := range kinds {
		images = append(images, knd.Image...)
	}

	return images
}
