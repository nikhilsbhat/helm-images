package pkg

import (
	"context"
	"errors"
	"os"
	"reflect"
	"regexp"
	"slices"
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
	ConfigMapImageRegex   = `\bimage\b`
	fetchingImagesMessage = "fetching the images from chart '%s' at path '%s'"
)

// Images represents GetImages.
type Images struct {
	Registries          []string   `json:"registries,omitempty"              yaml:"registries,omitempty"`
	Kind                []string   `json:"kind,omitempty"                    yaml:"kind,omitempty"`
	Values              []string   `json:"values,omitempty"                  yaml:"values,omitempty"`
	StringValues        []string   `json:"string_values,omitempty"           yaml:"string_values,omitempty"`
	FileValues          []string   `json:"file_values,omitempty"             yaml:"file_values,omitempty"`
	ShowOnly            []string   `json:"show_only,omitempty"               yaml:"show_only,omitempty"`
	Skip                []string   `json:"skip,omitempty"                    yaml:"skip,omitempty"`
	SkipReleases        []string   `json:"skip_releases,omitempty"           yaml:"skip_releases,omitempty"`
	Version             string     `json:"version,omitempty"                 yaml:"version,omitempty"`
	ImageRegex          string     `json:"image_regex,omitempty"             yaml:"image_regex,omitempty"`
	ConfigMapImageRegex string     `json:"configmap_image_regex,omitempty"   yaml:"configmap_image_regex,omitempty"`
	ValueFiles          ValueFiles `json:"value_files,omitempty"             yaml:"value_files,omitempty"`
	LogLevel            string     `json:"log_level,omitempty"               yaml:"log_level,omitempty"`
	OutputFormat        string     `json:"output_format,omitempty"           yaml:"output_format,omitempty"`
	ChartsDir           string     `json:"charts_dir,omitempty"              yaml:"charts_dir,omitempty"`
	Revision            int        `json:"revision,omitempty"                yaml:"revision,omitempty"`
	Raw                 bool       `json:"raw,omitempty"                     yaml:"raw,omitempty"`
	SkipTests           bool       `json:"skip_tests,omitempty"              yaml:"skip_tests,omitempty"`
	SkipCRDS            bool       `json:"skip_crds,omitempty"               yaml:"skip_crds,omitempty"`
	FromRelease         bool       `json:"from_release,omitempty"            yaml:"from_release,omitempty"`
	UniqueImages        bool       `json:"unique_images,omitempty"           yaml:"unique_images,omitempty"`
	NoColor             bool       `json:"no_color,omitempty"                yaml:"no_color,omitempty"`
	Validate            bool       `json:"validate,omitempty"                yaml:"validate,omitempty"`
	IsDefaultNamespace  bool       `json:"is_default_namespace,omitempty"    yaml:"is_default_namespace,omitempty"`
	Quiet               bool       `json:"quiet,omitempty"                   yaml:"quiet,omitempty"`
	releasesToSkip      []skipReleaseInfo
	json                bool
	yaml                bool
	table               bool
	csv                 bool
	all                 bool
	raw                 []byte
	release             string
	chart               string
	namespace           string
	log                 *logrus.Logger
	renderer            renderer.Config
}

type Skip struct {
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`
}

// SetAll would be set when images to be retrieved from all releases.
func (image *Images) SetAll(all bool) {
	image.all = all
}

// SetNamespace sets namespace passed.
func (image *Images) SetNamespace(namespace string) {
	image.namespace = namespace
}

// SetRelease sets release passed.
func (image *Images) SetRelease(release string) {
	image.release = release
}

// SetChart sets chart passed.
func (image *Images) SetChart(chart string) {
	image.chart = chart
}

// SetChartsDir sets charts directory passed.
func (image *Images) SetChartsDir(chartsDir string) {
	image.ChartsDir = chartsDir
}

// GetChartsDir returns the charts directory set under Images.
func (image *Images) GetChartsDir() string {
	return image.ChartsDir
}

// SetRaw sets raw.
func (image *Images) SetRaw(raw []byte) {
	image.raw = raw
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

// GetNamespace returns the namespace set under Images.
func (image *Images) GetNamespace() string {
	return image.namespace
}

// GetChart returns the chart set under Images.
func (image *Images) GetChart() string {
	return image.chart
}

// GetImages fetches all available images from the specified chart.
// Also filters identified images, to get just unique ones.
func (image *Images) GetImages(ctx context.Context) error {
	image.log.Debugf("got all required values to fetch the images from chart/release '%s' proceeding furter to fetch the same", image.release)

	chart, err := image.getChartManifests(ctx)
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
	var (
		img *k8s.Image
		err error
	)

	switch currentKind {
	case k8s.KindDeployment:
		img, err = k8s.NewDeployment().Get(kubeKindTemplate, "", image.log)
	case k8s.KindStatefulSet:
		img, err = k8s.NewStatefulSet().Get(kubeKindTemplate, "", image.log)
	case k8s.KindDaemonSet:
		img, err = k8s.NewDaemonSet().Get(kubeKindTemplate, "", image.log)
	case k8s.KindReplicaSet:
		img, err = k8s.NewReplicaSets().Get(kubeKindTemplate, "", image.log)
	case k8s.KindPod:
		img, err = k8s.NewPod().Get(kubeKindTemplate, "", image.log)
	case k8s.KindConfigMap:
		img, err = k8s.NewConfigMap().Get(kubeKindTemplate, image.ConfigMapImageRegex, image.log)
	case k8s.KindCronJob:
		img, err = k8s.NewCronjob().Get(kubeKindTemplate, "", image.log)
	case k8s.KindJob:
		img, err = k8s.NewJob().Get(kubeKindTemplate, "", image.log)
	case monitoringV1.AlertmanagersKind:
		img, err = k8s.NewAlertManager().Get(kubeKindTemplate, "", image.log)
	case monitoringV1.PrometheusesKind:
		img, err = k8s.NewPrometheus().Get(kubeKindTemplate, "", image.log)
	case monitoringV1.ThanosRulerKind:
		img, err = k8s.NewThanosRuler().Get(kubeKindTemplate, "", image.log)
	case k8s.KindThanos:
		img, err = k8s.NewThanos().Get(kubeKindTemplate, "", image.log)
	case k8s.KindThanosReceiver:
		img, err = k8s.NewThanosReceiver().Get(kubeKindTemplate, "", image.log)
	case k8s.KindGrafana:
		img, err = k8s.NewGrafana().Get(kubeKindTemplate, "", image.log)
	case k8s.KindCrossPlaneProvider:
		img, err = k8s.NewCrossPlaneProvider().Get(kubeKindTemplate, "", image.log)
	case k8s.KindCrossPlaneConfiguration:
		img, err = k8s.NewCrossPlaneConfiguration().Get(kubeKindTemplate, "", image.log)
	case k8s.KindCrossPlaneFunction:
		img, err = k8s.NewCrossPlaneFunction().Get(kubeKindTemplate, "", image.log)
	default:
		image.log.Debugf("kind '%s' is not supported at the moment", currentKind)

		return nil, nil
	}

	if err != nil {
		var grafanaErr *imgErrors.GrafanaAPIVersionSupportError

		if currentKind == k8s.KindGrafana && errors.As(err, &grafanaErr) {
			image.log.Errorf("fetching images from Kind Grafana errored with %s", err.Error())

			return nil, nil
		}

		return nil, err
	}

	if currentKind == k8s.KindConfigMap && reflect.DeepEqual(img, &k8s.Image{}) {
		return nil, nil
	}

	return []*k8s.Image{img}, nil
}

// GetImagesFromKind returns list of images from array of k8s.Image.
func GetImagesFromKind(kinds []*k8s.Image) []string {
	var images []string

	for _, knd := range kinds {
		images = append(images, knd.Image...)
	}

	return images
}

// GetImagesFromChartsDir fetches images from all helm charts in the specified directory.
func (image *Images) GetImagesFromChartsDir(ctx context.Context) error {
	charts, err := image.getChartsFromDir()
	if err != nil {
		return err
	}

	if image.isSimpleOutput() {
		return image.renderSimpleChartOutput(ctx, charts)
	}

	return image.renderStructuredChartOutput(ctx, charts)
}

func (image *Images) getChartManifests(ctx context.Context) ([]byte, error) {
	if image.Raw {
		image.log.Debug("reading the manifest from stdin")

		return image.raw, nil
	}

	if image.FromRelease {
		image.log.Debugf("from-release is selected, hence fetching manifests for '%s' from helm release", image.release)

		return image.getChartFromRelease()
	}

	image.log.Debugf("fetching manifests for '%s' by rendering helm template locally", image.release)

	return image.getChartFromTemplate(ctx)
}

func (image *Images) isSimpleOutput() bool {
	return !image.json && !image.yaml && !image.table && !image.csv
}

func (image *Images) renderSimpleChartOutput(ctx context.Context, charts []chartInfo) error {
	allImages := make([]string, 0)

	for _, chart := range charts {
		images, err := image.collectImagesFromChart(ctx, chart)
		if err != nil {
			return err
		}

		if len(images) == 0 {
			image.log.Infof("the chart '%s' does not have any images", chart.name)

			continue
		}

		images = image.FilterImagesByRegistriesNew(images)
		imageNames := GetImagesFromKind(images)
		allImages = append(allImages, imageNames...)
	}

	if image.UniqueImages {
		allImages = GetUniqEntries(allImages)
	}

	return image.renderer.Render(strings.Join(allImages, "\n"))
}

func (image *Images) renderStructuredChartOutput(ctx context.Context, charts []chartInfo) error {
	imagesFromAllCharts := make([]k8s.Images, 0)

	for _, chart := range charts {
		images, err := image.collectImagesFromChart(ctx, chart)
		if err != nil {
			return err
		}

		if len(images) == 0 {
			image.log.Infof("the chart '%s' does not have any images", chart.name)

			continue
		}

		output := image.setOutput(images)

		imagesFromAllCharts = append(imagesFromAllCharts, k8s.Images{
			ImagesFromRelease: output,
			NameSpace:         chart.name,
		})
	}

	return image.renderer.Render(imagesFromAllCharts)
}

func (image *Images) collectImagesFromChart(ctx context.Context, chart chartInfo) ([]*k8s.Image, error) {
	image.log.Debugf(fetchingImagesMessage, chart.name, chart.path)

	manifest, err := image.getChartManifestFromDir(ctx, chart.path, chart.name)
	if err != nil {
		image.log.Errorf("failed to render chart '%s': %v", chart.name, err)

		return nil, nil
	}

	kubeKindTemplates := image.GetTemplates(manifest)
	skips := image.GetResourcesToSkip()
	images := make([]*k8s.Image, 0)

	for _, kubeKindTemplate := range kubeKindTemplates {
		currentManifestName, currentKind, err := image.getManifestMetadata(kubeKindTemplate)
		if err != nil {
			return nil, err
		}

		if !slices.Contains(image.Kind, currentKind) {
			image.log.Debugf("either helm-images plugin does not support kind '%s' "+
				"at the moment or manifest might not have images to filter", currentKind)

			continue
		}

		if image.shouldSkipResource(skips, currentManifestName, currentKind) {
			image.log.Debugf("Skipping '%s' bearing name '%s' since it is set to skip.", currentKind, currentManifestName)

			continue
		}

		image.log.Debugf("fetching images from '%s' of kind '%s'", currentKind, currentManifestName)

		imagesFound, err := image.GetImage(currentKind, kubeKindTemplate)
		if err != nil {
			return nil, err
		}

		images = append(images, imagesFound...)
	}

	return images, nil
}

func (image *Images) getManifestMetadata(kubeKindTemplate string) (string, string, error) {
	currentManifestName, err := k8s.NewName().Get(kubeKindTemplate, image.log)
	if err != nil {
		return "", "", err
	}

	currentKind, err := k8s.NewKind().Get(kubeKindTemplate, image.log)
	if err != nil {
		return "", "", err
	}

	return currentManifestName, currentKind, nil
}

func (image *Images) shouldSkipResource(skips []Skip, manifestName, kind string) bool {
	manifestName = strings.ToLower(manifestName)
	kind = strings.ToLower(kind)

	for _, skip := range skips {
		if skip.Name == manifestName && skip.Kind == kind {
			return true
		}
	}

	return false
}
