package pkg

import (
	"bufio"
	"errors"
	"io"
	"regexp"

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
	Registries   []string   `json:"registries,omitempty"    yaml:"registries,omitempty"`
	Kind         []string   `json:"kind,omitempty"          yaml:"kind,omitempty"`
	Values       []string   `json:"values,omitempty"        yaml:"values,omitempty"`
	StringValues []string   `json:"string_values,omitempty" yaml:"string_values,omitempty"`
	FileValues   []string   `json:"file_values,omitempty"   yaml:"file_values,omitempty"`
	ImageRegex   string     `json:"image_regex,omitempty"   yaml:"image_regex,omitempty"`
	ValueFiles   ValueFiles `json:"value_files,omitempty"   yaml:"value_files,omitempty"`
	LogLevel     string     `json:"log_level,omitempty"     yaml:"log_level,omitempty"`
	SkipTests    bool       `json:"skip_tests,omitempty"    yaml:"skip_tests,omitempty"`
	SkipCRDS     bool       `json:"skip_crds,omitempty"     yaml:"skip_crds,omitempty"`
	FromRelease  bool       `json:"from_release,omitempty"  yaml:"from_release,omitempty"`
	UniqueImages bool       `json:"unique_images,omitempty" yaml:"unique_images,omitempty"`
	JSON         bool       `json:"json,omitempty"          yaml:"json,omitempty"`
	YAML         bool       `json:"yaml,omitempty"          yaml:"yaml,omitempty"`
	Table        bool       `json:"table,omitempty"         yaml:"table,omitempty"`
	NoColor      bool       `json:"no_color,omitempty"      yaml:"no_color,omitempty"`
	release      string
	chart        string
	log          *logrus.Logger
	writer       *bufio.Writer
}

// SetRelease sets release passed.
func (image *Images) SetRelease(release string) {
	image.release = release
}

// SetChart sets chart passed.
func (image *Images) SetChart(chart string) {
	image.chart = chart
}

// SetWriter sets writer to Images.
func (image *Images) SetWriter(writer io.Writer) {
	image.writer = bufio.NewWriter(writer)
}

// GetRelease returns the release set under Images.
func (image *Images) GetRelease() string {
	return image.release
}

// GetChart returns the chart set under Images.
func (image *Images) GetChart() string {
	return image.chart
}

// GetWriter returns the writer set under Images.
func (image *Images) GetWriter() *bufio.Writer {
	return image.writer
}

// GetImages fetches all available images from the specified chart.
// Also filters identified images, to get just unique ones.
//
//nolint:funlen,gocognit
func (image *Images) GetImages() error {
	image.log.Debugf("got all required values to fetch the images from chart/release '%s' proceeding furter to fetch the same", image.release)

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
			image.log.Debugf("either helm-images plugin does not support kind '%s' "+
				"at the moment or manifest might not have images to filter", currentKind)

			continue
		}

		image.log.Debugf("fetching images from kind '%s'", currentKind)

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
					image.log.Errorf("fetching images from Kind Grafana errored with %s", err.Error())

					continue
				}

				return err
			}

			images = append(images, grafana)
		default:
			image.log.Debugf("kind '%s' is not supported at the moment", currentKind)
		}
	}

	if len(images) == 0 {
		switch image.FromRelease {
		case true:
			image.log.Infof("the release '%s' does not have any images", image.release)

			return nil
		default:
			image.log.Infof("the chart '%s' does not have any images", image.chart)

			return nil
		}
	}

	return image.render(images)
}

func (image *Images) getChartManifests() ([]byte, error) {
	if image.FromRelease {
		image.log.Debugf("from-release is selected, hence fetching manifests for '%s' from helm release", image.release)

		return image.getChartFromRelease()
	}

	image.log.Debugf("fetching manifests for '%s' by rendering helm template locally", image.release)

	return image.getChartFromTemplate()
}

// GetTemplates returns the split manifests fetched from one big template string fetched from `helm template`.
func (image *Images) GetTemplates(template []byte) []string {
	image.log.Debugf("splitting helm manifests with regex pattern: '%s'", image.ImageRegex)
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
