package pkg

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"regexp"
	"strings"

	imgErrors "github.com/nikhilsbhat/helm-images/pkg/errors"
	"github.com/nikhilsbhat/helm-images/pkg/k8s"
	monitoringV1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
)

const (
	// ImageRegex is the default regex, that is used to split one big helm template to multiple templates.
	// Splitting templates eases the task of  identifying Kubernetes objects.
	ImageRegex = `---\n# Source:\s.*.`
)

// Images represents GetImages.
type Images struct {
	// Registries are list of registry names which we have filter out from
	Registries   []string
	Kind         []string
	Values       []string
	StringValues []string
	FileValues   []string
	ImageRegex   string
	ValueFiles   ValueFiles
	LogLevel     string
	SkipTests    bool
	SkipCRDS     bool
	FromRelease  bool
	UniqueImages bool
	JSON         bool
	YAML         bool
	Table        bool
	release      string
	chart        string
	log          *logrus.Logger
	writer       *bufio.Writer
}

func (image *Images) SetRelease(release string) {
	image.release = release
}

func (image *Images) SetChart(chart string) {
	image.chart = chart
}

func (image *Images) SetWriter(writer io.Writer) {
	image.writer = bufio.NewWriter(writer)
}

func (image *Images) GetRelease() string {
	return image.release
}

func (image *Images) GetChart() string {
	return image.chart
}

func (image *Images) GetWriter() *bufio.Writer {
	return image.writer
}

// GetImages fetches all available images from the specified chart.
// Also filters identified images, to get just unique ones.
//
//nolint:funlen,gocognit
func (image *Images) GetImages() error {
	image.log.Debug(
		fmt.Sprintf("got all required values to fetch the images from chart/release '%s' proceeding furter to fetch the same", image.release),
	)

	chart, err := image.getChartManifests()
	if err != nil {
		return err
	}

	images := make([]*k8s.Image, 0)
	kubeKindTemplates := image.GetTemplates(chart)

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
		case monitoringV1.AlertmanagersKind:
			alertManager, err := k8s.NewAlertManager().Get(kubeKindTemplate)
			if err != nil {
				return err
			}

			images = append(images, alertManager)
		case monitoringV1.PrometheusesKind:
			prometheus, err := k8s.NewPrometheus().Get(kubeKindTemplate)
			if err != nil {
				return err
			}

			images = append(images, prometheus)
		case monitoringV1.ThanosRulerKind:
			thanosRuler, err := k8s.NewThanosRuler().Get(kubeKindTemplate)
			if err != nil {
				return err
			}

			images = append(images, thanosRuler)
		case k8s.KindThanos:
			thanos, err := k8s.NewThanos().Get(kubeKindTemplate)
			if err != nil {
				return err
			}

			images = append(images, thanos)
		case k8s.KindThanosReceiver:
			thanosReceiver, err := k8s.NewThanosReceiver().Get(kubeKindTemplate)
			if err != nil {
				return err
			}

			images = append(images, thanosReceiver)
		case k8s.KindGrafana:
			grafana, err := k8s.NewGrafana().Get(kubeKindTemplate)

			grafanaErr := &imgErrors.GrafanaAPIVersionSupportError{}
			if err != nil {
				if errors.As(err, &grafanaErr) {
					image.log.Debugf("fetching images from Kind Grafana errored with %s", err.Error())

					continue
				}

				return err
			}

			images = append(images, grafana)
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

	if strings.ToLower(image.LogLevel) == logrus.DebugLevel.String() {
		flags = append(flags, "--debug")
	}

	if image.SkipTests {
		flags = append(flags, "--skip-tests")
	}

	if image.SkipCRDS {
		flags = append(flags, "--skip-crds")
	}

	args := []string{"template", image.release, image.chart}
	args = append(args, flags...)

	image.log.Debug(fmt.Sprintf("rendering helm chart with following commands/flags '%s'", strings.Join(args, ", ")))

	cmd := exec.Command(os.Getenv("HELM_BIN"), args...) //nolint:gosec
	output, err := cmd.Output()

	var exitErr *exec.ExitError

	if errors.As(err, &exitErr) {
		image.log.Errorf("rendering template for release: '%s' errored with %v", image.release, err)

		return nil, fmt.Errorf("%w: %s", exitErr, exitErr.Stderr)
	}

	var pathErr *fs.PathError

	if errors.As(err, &pathErr) {
		image.log.Error("locating helm cli errored with", err)

		return nil, fmt.Errorf("%w: %s", pathErr, pathErr.Path)
	}

	return output, nil
}

func (image *Images) GetTemplates(template []byte) []string {
	image.log.Debug(fmt.Sprintf("splitting helm manifests with regex pattern: '%s'", image.ImageRegex))
	temp := regexp.MustCompile(image.ImageRegex)
	kinds := temp.Split(string(template), -1)
	// Removing empty string at the beginning as splitting string always adds it in front.
	kinds = kinds[1:]

	return kinds
}

func GetImagesFromKind(kinds []*k8s.Image) []string {
	var images []string
	for _, knd := range kinds {
		images = append(images, knd.Image...)
	}

	return images
}
