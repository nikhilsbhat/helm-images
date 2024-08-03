package pkg

import (
	"errors"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/nikhilsbhat/common/renderer"
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
	// ConfigMapImageRegex is the default regex, that is used for identifying images from ConfigMap.
	ConfigMapImageRegex = `\bimage\b`
)

// Images represents GetImages.
type Images struct {
	Registries          []string   `json:"registries,omitempty"            yaml:"registries,omitempty"`
	Kind                []string   `json:"kind,omitempty"                  yaml:"kind,omitempty"`
	Values              []string   `json:"values,omitempty"                yaml:"values,omitempty"`
	StringValues        []string   `json:"string_values,omitempty"         yaml:"string_values,omitempty"`
	FileValues          []string   `json:"file_values,omitempty"           yaml:"file_values,omitempty"`
	ShowOnly            []string   `json:"show_only,omitempty"             yaml:"show_only,omitempty"`
	Skip                []string   `json:"skip,omitempty"                  yaml:"skip,omitempty"`
	Version             string     `json:"version,omitempty"               yaml:"version,omitempty"`
	ImageRegex          string     `json:"image_regex,omitempty"           yaml:"image_regex,omitempty"`
	ConfigMapImageRegex string     `json:"configmap_image_regex,omitempty" yaml:"configmap_image_regex,omitempty"`
	ValueFiles          ValueFiles `json:"value_files,omitempty"           yaml:"value_files,omitempty"`
	LogLevel            string     `json:"log_level,omitempty"             yaml:"log_level,omitempty"`
	OutputFormat        string     `json:"output_format,omitempty"         yaml:"output_format,omitempty"`
	Revision            int        `json:"revision,omitempty"              yaml:"revision,omitempty"`
	SkipTests           bool       `json:"skip_tests,omitempty"            yaml:"skip_tests,omitempty"`
	SkipCRDS            bool       `json:"skip_crds,omitempty"             yaml:"skip_crds,omitempty"`
	FromRelease         bool       `json:"from_release,omitempty"          yaml:"from_release,omitempty"`
	UniqueImages        bool       `json:"unique_images,omitempty"         yaml:"unique_images,omitempty"`
	NoColor             bool       `json:"no_color,omitempty"              yaml:"no_color,omitempty"`
	Validate            bool       `json:"validate,omitempty"              yaml:"validate,omitempty"`
	json                bool
	yaml                bool
	table               bool
	csv                 bool
	release             string
	chart               string
	log                 *logrus.Logger
	renderer            renderer.Config
}

type Skip struct {
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`
}

// SetRelease sets release passed.
func (image *Images) SetRelease(release string) {
	image.release = release
}

// SetChart sets chart passed.
func (image *Images) SetChart(chart string) {
	image.chart = chart
}

// SetRenderer sets renderer to Images.
func (image *Images) SetRenderer() {
	render := renderer.GetRenderer(os.Stdout, image.log, image.NoColor, image.yaml, image.json, image.csv, image.table)
	image.renderer = render
}

// GetRelease returns the release set under Images.
func (image *Images) GetRelease() string {
	return image.release
}

// GetChart returns the chart set under Images.
func (image *Images) GetChart() string {
	return image.chart
}

// GetImages fetches all available images from the specified chart.
// Also filters identified images, to get just unique ones.
func (image *Images) GetImages() error {
	image.log.Debugf("got all required values to fetch the images from chart/release '%s' proceeding furter to fetch the same", image.release)

	chart, err := image.getChartManifests()
	if err != nil {
		return err
	}

	images := make([]*k8s.Image, 0)
	kubeKindTemplates := image.GetTemplates(chart)
	skips := image.GetResourcesToSkip()

	for _, kubeKindTemplate := range kubeKindTemplates {
		currentManifestName, err := k8s.NewName().Get(kubeKindTemplate, image.log)
		if err != nil {
			return err
		}

		currentKind, err := k8s.NewKind().Get(kubeKindTemplate, image.log)
		if err != nil {
			return err
		}

		if !funk.Contains(image.Kind, currentKind) {
			image.log.Debugf("either helm-images plugin does not support kind '%s' "+
				"at the moment or manifest might not have images to filter", currentKind)

			continue
		}

		shouldSkip := false

		for _, skip := range skips {
			if skip.Name == strings.ToLower(currentManifestName) && skip.Kind == strings.ToLower(currentKind) {
				image.log.Debugf("Skipping '%s' bearing name '%s' since it is set to skip.", currentKind, currentManifestName)

				shouldSkip = true

				break
			}
		}

		if shouldSkip {
			continue
		}

		image.log.Debugf("fetching images from '%s' of kind '%s'", currentKind, currentManifestName)

		imagesFound, err := image.GetImage(currentKind, kubeKindTemplate)
		if err != nil {
			return err
		}

		images = append(images, imagesFound...)
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

	output := image.setOutput(images)

	return image.renderer.Render(output)
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

// GetResourcesToSkip returns the skip from translating the flags.
func (image *Images) GetResourcesToSkip() []Skip {
	skip := make([]Skip, 0)

	for _, resource := range image.Skip {
		splitResource := strings.Split(resource, "=")

		skipsContentCount := 2
		if len(splitResource) != skipsContentCount {
			continue
		}

		skip = append(skip, Skip{
			Kind: strings.ToLower(splitResource[0]),
			Name: strings.ToLower(splitResource[1]),
		})
	}

	return skip
}

// GetImage returns []*k8s.Image from the kubernetes manifests.
//
//nolint:gocognit,funlen
func (image *Images) GetImage(currentKind, kubeKindTemplate string) ([]*k8s.Image, error) {
	images := make([]*k8s.Image, 0)

	switch currentKind {
	case k8s.KindDeployment:
		deployImages, err := k8s.NewDeployment().Get(kubeKindTemplate, "", image.log)
		if err != nil {
			return nil, err
		}

		images = append(images, deployImages)
	case k8s.KindStatefulSet:
		stsImages, err := k8s.NewStatefulSet().Get(kubeKindTemplate, "", image.log)
		if err != nil {
			return nil, err
		}

		images = append(images, stsImages)
	case k8s.KindDaemonSet:
		daemonImages, err := k8s.NewDaemonSet().Get(kubeKindTemplate, "", image.log)
		if err != nil {
			return nil, err
		}

		images = append(images, daemonImages)
	case k8s.KindReplicaSet:
		replicaSets, err := k8s.NewReplicaSets().Get(kubeKindTemplate, "", image.log)
		if err != nil {
			return nil, err
		}

		images = append(images, replicaSets)
	case k8s.KindPod:
		pods, err := k8s.NewPod().Get(kubeKindTemplate, "", image.log)
		if err != nil {
			return nil, err
		}

		images = append(images, pods)
	case k8s.KindConfigMap:
		configMap, err := k8s.NewConfigMap().Get(kubeKindTemplate, image.ConfigMapImageRegex, image.log)
		if err != nil {
			return nil, err
		}

		if !reflect.DeepEqual(configMap, &k8s.Image{}) {
			images = append(images, configMap)
		}
	case k8s.KindCronJob:
		cronJob, err := k8s.NewCronjob().Get(kubeKindTemplate, "", image.log)
		if err != nil {
			return nil, err
		}

		images = append(images, cronJob)
	case k8s.KindJob:
		job, err := k8s.NewJob().Get(kubeKindTemplate, "", image.log)
		if err != nil {
			return nil, err
		}

		images = append(images, job)
	case monitoringV1.AlertmanagersKind:
		alertManager, err := k8s.NewAlertManager().Get(kubeKindTemplate, "", image.log)
		if err != nil {
			return nil, err
		}

		images = append(images, alertManager)
	case monitoringV1.PrometheusesKind:
		prometheus, err := k8s.NewPrometheus().Get(kubeKindTemplate, "", image.log)
		if err != nil {
			return nil, err
		}

		images = append(images, prometheus)
	case monitoringV1.ThanosRulerKind:
		thanosRuler, err := k8s.NewThanosRuler().Get(kubeKindTemplate, "", image.log)
		if err != nil {
			return nil, err
		}

		images = append(images, thanosRuler)
	case k8s.KindThanos:
		thanos, err := k8s.NewThanos().Get(kubeKindTemplate, "", image.log)
		if err != nil {
			return nil, err
		}

		images = append(images, thanos)
	case k8s.KindThanosReceiver:
		thanosReceiver, err := k8s.NewThanosReceiver().Get(kubeKindTemplate, "", image.log)
		if err != nil {
			return nil, err
		}

		images = append(images, thanosReceiver)
	case k8s.KindGrafana:
		grafana, err := k8s.NewGrafana().Get(kubeKindTemplate, "", image.log)

		grafanaErr := &imgErrors.GrafanaAPIVersionSupportError{}
		if err != nil {
			if errors.As(err, &grafanaErr) {
				image.log.Errorf("fetching images from Kind Grafana errored with %s", err.Error())

				return nil, nil
			}

			return nil, err
		}

		images = append(images, grafana)
	default:
		image.log.Debugf("kind '%s' is not supported at the moment", currentKind)
	}

	return images, nil
}

// GetImagesFromKind returns list of images from array of k8s.Image.
func GetImagesFromKind(kinds []*k8s.Image) []string {
	var images []string

	for _, knd := range kinds {
		images = append(images, knd.Image...)
	}

	return images
}
