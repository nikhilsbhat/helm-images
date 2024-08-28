package k8s

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	thanosAlphaV1 "github.com/banzaicloud/thanos-operator/pkg/sdk/api/v1alpha1"
	"github.com/ghodss/yaml"
	grafanaBetaV1 "github.com/grafana-operator/grafana-operator/api/v1beta1"
	"github.com/nikhilsbhat/common/content"
	imgErrors "github.com/nikhilsbhat/helm-images/pkg/errors"
	monitoringV1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	appsV1 "k8s.io/api/apps/v1"
	batchV1 "k8s.io/api/batch/v1"
	coreV1 "k8s.io/api/core/v1"
)

const (
	KindDeployment     = "Deployment"
	KindStatefulSet    = "StatefulSet"
	KindDaemonSet      = "DaemonSet"
	KindCronJob        = "CronJob"
	KindJob            = "Job"
	KindReplicaSet     = "ReplicaSet"
	KindPod            = "Pod"
	KindGrafana        = "Grafana"
	KindThanos         = "Thanos"
	KindThanosReceiver = "Receiver"
	KindConfigMap      = "ConfigMap"
	kubeKind           = "kind"
)

var imagesFlags = []string{
	"--prometheus-config-reloader",
	"--thanos-default-base-image",
	"--acme-http01-solver-image",
}

type (
	Deployments  appsV1.Deployment
	StatefulSets appsV1.StatefulSet
	DaemonSets   appsV1.DaemonSet
	ReplicaSets  appsV1.ReplicaSet
	CronJob      batchV1.CronJob
	Job          batchV1.Job
	Pod          coreV1.Pod
	Kind         map[string]interface{}
	Name         map[string]interface{}
	Resource     map[string]interface{}
	containers   struct {
		containers []coreV1.Container
	}
	AlertManager   monitoringV1.Alertmanager
	Prometheus     monitoringV1.Prometheus
	ThanosRuler    monitoringV1.ThanosRuler
	Grafana        grafanaBetaV1.Grafana
	Thanos         thanosAlphaV1.Thanos
	ThanosReceiver thanosAlphaV1.Receiver
	ConfigMap      coreV1.ConfigMap
)

// KindInterface implements method that identifies the type of kubernetes workloads.
type KindInterface interface {
	Get(dataMap map[string]interface{}, log *logrus.Logger) (string, error)
}

// ImagesInterface implements method that gets images from various kubernetes workloads.
type ImagesInterface interface {
	Get(dataMap map[string]interface{}, imageRegex string, log *logrus.Logger) (*Image, error)
}

// Image holds information of images retrieved.
type Image struct {
	Kind  string   `json:"kind,omitempty"  yaml:"kind,omitempty"`
	Name  string   `json:"name,omitempty"  yaml:"name,omitempty"`
	Image []string `json:"image,omitempty" yaml:"image,omitempty"`
}

type Images struct {
	ImagesFromRelease interface{} `json:"images_from_release,omitempty" yaml:"images_from_release,omitempty"`
	NameSpace         string      `json:"name_space,omitempty"          yaml:"name_space,omitempty"`
}

func (name *Name) Get(dataMap map[string]interface{}, log *logrus.Logger) (string, error) {
	kindYaml := *name

	metadata, metadataExists := kindYaml["metadata"].(map[string]interface{})
	if !metadataExists {
		log.Warn("failed to get 'metadata' from the manifest")

		return "", nil
	}

	metadataName, metadataNameExists := metadata["name"].(string)
	if !metadataNameExists {
		return "", &imgErrors.ImageError{Message: "failed to get name from the manifest, 'name' is not type string"}
	}

	return metadataName, nil
}

func (kin *Kind) Get(dataMap map[string]interface{}, log *logrus.Logger) (string, error) {
	if kin == nil {
		log.Warn("looks like it manifest is empty")

		return "", nil
	}

	kindYaml := *kin

	kind, kindExists := kindYaml[kubeKind].(string)
	if !kindExists {
		log.Warn("failed to get 'kind' from the manifest")
		return "", nil
	}

	return kind, nil
}

// Get identifies images from Deployments.
func (dep *Deployments) Get(dataMap map[string]interface{}, _ string, _ *logrus.Logger) (*Image, error) {
	depContainers := containers{append(dep.Spec.Template.Spec.Containers, dep.Spec.Template.Spec.InitContainers...)}

	images := &Image{
		Kind:  KindDeployment,
		Name:  dep.Name,
		Image: depContainers.getImages(),
	}

	images.Image = append(images.Image, depContainers.getImagesFromArgs()...)

	return images, nil
}

// Get identifies images from StatefulSets.
func (dep *StatefulSets) Get(dataMap map[string]interface{}, _ string, _ *logrus.Logger) (*Image, error) {
	depContainers := containers{append(dep.Spec.Template.Spec.Containers, dep.Spec.Template.Spec.InitContainers...)}

	images := &Image{
		Kind:  KindStatefulSet,
		Name:  dep.Name,
		Image: depContainers.getImages(),
	}

	images.Image = append(images.Image, depContainers.getImagesFromArgs()...)

	return images, nil
}

// Get identifies images from DaemonSets.
func (dep *DaemonSets) Get(dataMap map[string]interface{}, _ string, _ *logrus.Logger) (*Image, error) {
	depContainers := containers{append(dep.Spec.Template.Spec.Containers, dep.Spec.Template.Spec.InitContainers...)}

	images := &Image{
		Kind:  KindDaemonSet,
		Name:  dep.Name,
		Image: depContainers.getImages(),
	}

	images.Image = append(images.Image, depContainers.getImagesFromArgs()...)

	return images, nil
}

// Get identifies images from CronJob.
func (dep *CronJob) Get(dataMap map[string]interface{}, _ string, _ *logrus.Logger) (*Image, error) {
	depContainers := containers{append(dep.Spec.JobTemplate.Spec.Template.Spec.Containers,
		dep.Spec.JobTemplate.Spec.Template.Spec.InitContainers...)}

	images := &Image{
		Kind:  KindCronJob,
		Name:  dep.Name,
		Image: depContainers.getImages(),
	}

	images.Image = append(images.Image, depContainers.getImagesFromArgs()...)

	return images, nil
}

// Get identifies images from Job.
func (dep *Job) Get(dataMap map[string]interface{}, _ string, _ *logrus.Logger) (*Image, error) {
	depContainers := containers{append(dep.Spec.Template.Spec.Containers, dep.Spec.Template.Spec.InitContainers...)}

	images := &Image{
		Kind:  KindJob,
		Name:  dep.Name,
		Image: depContainers.getImages(),
	}

	images.Image = append(images.Image, depContainers.getImagesFromArgs()...)

	return images, nil
}

// Get identifies images from ReplicaSets.
func (dep *ReplicaSets) Get(dataMap map[string]interface{}, _ string, _ *logrus.Logger) (*Image, error) {
	depContainers := containers{append(dep.Spec.Template.Spec.Containers, dep.Spec.Template.Spec.InitContainers...)}

	images := &Image{
		Kind:  KindReplicaSet,
		Name:  dep.Name,
		Image: depContainers.getImages(),
	}

	images.Image = append(images.Image, depContainers.getImagesFromArgs()...)

	return images, nil
}

// Get identifies images from Pod.
func (dep *Pod) Get(dataMap map[string]interface{}, _ string, _ *logrus.Logger) (*Image, error) {
	depContainers := containers{append(dep.Spec.Containers, dep.Spec.InitContainers...)}

	images := &Image{
		Kind:  KindPod,
		Name:  dep.Name,
		Image: depContainers.getImages(),
	}

	images.Image = append(images.Image, depContainers.getImagesFromArgs()...)

	return images, nil
}

// Get identifies images from AlertManager.
func (dep *AlertManager) Get(dataMap map[string]interface{}, _ string, _ *logrus.Logger) (*Image, error) {
	images := &Image{
		Kind:  monitoringV1.AlertmanagersKind,
		Name:  dep.Name,
		Image: []string{*dep.Spec.Image},
	}

	return images, nil
}

// Get identifies images from Prometheus.
func (dep *Prometheus) Get(dataMap map[string]interface{}, _ string, _ *logrus.Logger) (*Image, error) {
	var imageNames []string

	depContainers := containers{append(dep.Spec.Containers, dep.Spec.InitContainers...)}

	imageNames = append(imageNames, depContainers.getImages()...)
	imageNames = append(imageNames, *dep.Spec.Image)

	images := &Image{
		Kind:  monitoringV1.PrometheusesKind,
		Name:  dep.Name,
		Image: imageNames,
	}

	return images, nil
}

// Get identifies images from ThanosRuler.
func (dep *ThanosRuler) Get(dataMap map[string]interface{}, _ string, _ *logrus.Logger) (*Image, error) {
	var imageNames []string

	depContainers := containers{append(dep.Spec.Containers, dep.Spec.InitContainers...)}

	imageNames = append(imageNames, depContainers.getImages()...)
	imageNames = append(imageNames, dep.Spec.Image)

	images := &Image{
		Kind:  monitoringV1.ThanosRulerKind,
		Name:  dep.Name,
		Image: imageNames,
	}

	return images, nil
}

// Get identifies images from Grafana.
func (dep *Grafana) Get(dataMap map[string]interface{}, _ string, _ *logrus.Logger) (*Image, error) {
	if dep.APIVersion == "integreatly.org/v1alpha1" {
		return nil, &imgErrors.GrafanaAPIVersionSupportError{
			Message: fmt.Sprintf("plugin supports the latest api version and '%s' is not supported", dep.APIVersion),
		}
	}

	grafanaDeployment := dep.Spec.Deployment.Spec.Template.Spec
	depContainers := containers{append(grafanaDeployment.Containers, grafanaDeployment.InitContainers...)}

	images := &Image{
		Kind:  KindGrafana,
		Name:  dep.Name,
		Image: depContainers.getImages(),
	}

	return images, nil
}

// Get identifies images from Thanos.
func (dep *Thanos) Get(dataMap map[string]interface{}, _ string, _ *logrus.Logger) (*Image, error) {
	thanosContainers := make([]coreV1.Container, 0)
	thanosContainers = append(thanosContainers, dep.Spec.Rule.StatefulsetOverrides.Spec.Template.Spec.Containers...)
	thanosContainers = append(thanosContainers, dep.Spec.Rule.StatefulsetOverrides.Spec.Template.Spec.InitContainers...)
	thanosContainers = append(thanosContainers, dep.Spec.Query.DeploymentOverrides.Spec.Template.Spec.Containers...)
	thanosContainers = append(thanosContainers, dep.Spec.Query.DeploymentOverrides.Spec.Template.Spec.InitContainers...)
	thanosContainers = append(thanosContainers, dep.Spec.StoreGateway.DeploymentOverrides.Spec.Template.Spec.Containers...)
	thanosContainers = append(thanosContainers, dep.Spec.StoreGateway.DeploymentOverrides.Spec.Template.Spec.InitContainers...)
	thanosContainers = append(thanosContainers, dep.Spec.QueryFrontend.DeploymentOverrides.Spec.Template.Spec.Containers...)
	thanosContainers = append(thanosContainers, dep.Spec.QueryFrontend.DeploymentOverrides.Spec.Template.Spec.InitContainers...)

	depContainers := containers{thanosContainers}

	images := &Image{
		Kind:  KindThanos,
		Name:  dep.Name,
		Image: depContainers.getImages(),
	}

	return images, nil
}

// Get identifies images from ThanosReceiver.
func (dep *ThanosReceiver) Get(dataMap map[string]interface{}, _ string, _ *logrus.Logger) (*Image, error) {
	receiverGroupTotalContainers := make([]coreV1.Container, 0)

	for _, receiverGroup := range dep.Spec.ReceiverGroups {
		receiverGroupTotalContainers = append(receiverGroupTotalContainers, receiverGroup.StatefulSetOverrides.Spec.Template.Spec.Containers...)
		receiverGroupTotalContainers = append(receiverGroupTotalContainers,
			receiverGroup.StatefulSetOverrides.Spec.Template.Spec.InitContainers...)
	}

	depContainers := containers{receiverGroupTotalContainers}

	images := &Image{
		Kind:  KindThanosReceiver,
		Name:  dep.Name,
		Image: depContainers.getImages(),
	}

	return images, nil
}

func (dep *ConfigMap) Get(dataMap map[string]interface{}, imageRegex string, log *logrus.Logger) (*Image, error) {
	images := &Image{
		Kind:  KindConfigMap,
		Name:  dep.Name,
		Image: make([]string, 0),
	}

	log.Debugf("using regex '%s' for identifying images from configmap", imageRegex)

	for key, value := range dep.Data {
		var valueMap interface{}

		object := content.Object(value)

		switch objType := object.CheckFileType(log); objType {
		case content.FileTypeYAML:
			if err := yaml.Unmarshal([]byte(object.String()), &valueMap); err != nil {
				log.Errorf("deserializing yaml data of configmap '%s' errored with '%s'", dep.Name, err.Error())

				continue
			}

			valuesFound, found := GetImage(GetData(valueMap), key, imageRegex, log)
			if !found {
				continue
			}

			images.Image = append(images.Image, valuesFound...)
		case content.FileTypeJSON:
			if err := json.Unmarshal([]byte(object.String()), &valueMap); err != nil {
				log.Errorf("deserializing json data of configmap '%s' errored with '%s'", dep.Name, err.Error())

				continue
			}

			valuesFound, found := GetImage(GetData(valueMap), key, imageRegex, log)
			if !found {
				continue
			}

			images.Image = append(images.Image, valuesFound...)
		case content.FileTypeString, content.FileTypeUnknown:
			imageFound, err := imageMatch(imageRegex, strings.ToLower(key))
			if err != nil {
				return nil, err
			}

			if imageFound {
				images.Image = append(images.Image, value)
			}
		}
	}

	if len(images.Image) == 0 {
		return &Image{}, nil
	}

	return images, nil
}

// NewDeployment returns new instance of Deployments.
func NewDeployment() ImagesInterface {
	return &Deployments{}
}

// NewStatefulSet returns new instance of StatefulSets.
func NewStatefulSet() ImagesInterface {
	return &StatefulSets{}
}

// NewDaemonSet returns new instance of DaemonSets.
func NewDaemonSet() ImagesInterface {
	return &DaemonSets{}
}

// NewReplicaSets returns new instance of ReplicaSets.
func NewReplicaSets() ImagesInterface {
	return &ReplicaSets{}
}

// NewCronjob returns new instance of Cronjob.
func NewCronjob() ImagesInterface {
	return &CronJob{}
}

// NewJob returns new instance of Job.
func NewJob() ImagesInterface {
	return &Job{}
}

// NewPod returns new instance of Pod.
func NewPod() ImagesInterface {
	return &Pod{}
}

// NewAlertManager returns new instance of AlertManager.
func NewAlertManager() ImagesInterface {
	return &AlertManager{}
}

// NewPrometheus returns new instance of Prometheus.
func NewPrometheus() ImagesInterface {
	return &Prometheus{}
}

// NewThanosRuler returns new instance of ThanosRuler.
func NewThanosRuler() ImagesInterface {
	return &ThanosRuler{}
}

// NewGrafana returns new instance of Grafana.
func NewGrafana() ImagesInterface {
	return &Grafana{}
}

// NewThanos returns new instance of Thanos.
func NewThanos() ImagesInterface {
	return &Thanos{}
}

// NewThanosReceiver returns new instance of ThanosReceiver.
func NewThanosReceiver() ImagesInterface {
	return &ThanosReceiver{}
}

func NewConfigMap() ImagesInterface {
	return &ConfigMap{}
}

// NewKind returns new instance of Kind.
func NewKind() KindInterface {
	return &Kind{}
}

// NewName returns new instance of Name.
func NewName() KindInterface {
	return &Name{}
}

func SupportedKinds() []string {
	kinds := []string{
		KindDeployment, KindStatefulSet, KindDaemonSet,
		KindCronJob, KindJob, KindReplicaSet, KindPod,
		monitoringV1.AlertmanagersKind, monitoringV1.PrometheusesKind, monitoringV1.ThanosRulerKind,
		KindGrafana, KindThanos, KindThanosReceiver, KindConfigMap,
	}

	return kinds
}

// kube-prometheus-stack/prometheus-operator supplies config-reloader and thanos
// images through container args.
func (cont containers) getImagesFromArgs() []string {
	images := make([]string, 0)

	for _, container := range cont.containers {
		for _, arg := range container.Args {
			keyValue := strings.Split(arg, "=")
			if len(keyValue) == 2 && funk.Contains(imagesFlags, keyValue[0]) {
				images = append(images, keyValue[1])
			}
		}
	}

	return images
}

func (cont containers) getImages() []string {
	images := make([]string, 0)
	for _, container := range cont.containers {
		images = append(images, container.Image)
	}

	return images
}

//nolint:nonamedreturns
func GetImage(data map[string]any, key, regex string, log *logrus.Logger) (values []string, valuesFound bool) {
	for dataKey, dataValue := range data {
		imageFound, err := imageMatch(regex, strings.ToLower(dataKey))
		if err != nil {
			return nil, false
		}

		if imageFound {
			log.Debugf("found image '%s' for regex '%s'", dataKey, regex)

			if strValue, ok := dataValue.(string); ok && len(strValue) > 0 {
				values = append(values, strValue)
				valuesFound = true
			}
		}

		switch dataValueType := dataValue.(type) {
		case []interface{}:
			for _, item := range dataValueType {
				if nestedMap, ok := item.(map[string]interface{}); ok {
					if nestedValues, found := GetImage(nestedMap, key, regex, log); found {
						values = append(values, nestedValues...)
						valuesFound = true
					}
				}
			}
		case map[string]interface{}:
			if nestedValues, found := GetImage(dataValueType, key, regex, log); found {
				values = append(values, nestedValues...)
				valuesFound = true
			}
		}
	}

	return values, valuesFound
}

func GetData(value interface{}) map[string]interface{} {
	valueMap := make(map[string]interface{})

	switch dataValueType := value.(type) {
	case []map[string]interface{}:
		if len(dataValueType) > 0 {
			return dataValueType[0]
		}
	case map[string]interface{}:
		return dataValueType
	case []interface{}:
		for _, item := range dataValueType {
			if nestedMap, ok := item.(map[string]interface{}); ok {
				for nestedKey, nestedValue := range nestedMap {
					valueMap[nestedKey] = nestedValue
				}
			}
		}
	}

	return valueMap
}

func imageMatch(imageRegex, imageString string) (bool, error) {
	regex, err := regexp.Compile(imageRegex)
	if err != nil {
		return false, err
	}

	return regex.MatchString(imageString), nil
}
