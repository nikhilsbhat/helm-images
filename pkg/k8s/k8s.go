package k8s

import (
	"fmt"

	thanosAlphaV1 "github.com/banzaicloud/thanos-operator/pkg/sdk/api/v1alpha1"
	"github.com/ghodss/yaml"
	grafanaBetaV1 "github.com/grafana-operator/grafana-operator/api/v1beta1"
	imgErrors "github.com/nikhilsbhat/helm-images/pkg/errors"
	monitoringV1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
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
	kubeKind           = "kind"
)

type (
	Deployments  appsV1.Deployment
	StatefulSets appsV1.StatefulSet
	DaemonSets   appsV1.DaemonSet
	ReplicaSets  appsV1.ReplicaSet
	CronJob      batchV1.CronJob
	Job          batchV1.Job
	Pod          coreV1.Pod
	Kind         map[string]interface{}
	containers   struct {
		containers []coreV1.Container
	}
	AlertManager   monitoringV1.Alertmanager
	Prometheus     monitoringV1.Prometheus
	ThanosRuler    monitoringV1.ThanosRuler
	Grafana        grafanaBetaV1.Grafana
	Thanos         thanosAlphaV1.Thanos
	ThanosReceiver thanosAlphaV1.Receiver
)

// KindInterface implements method that identifies the type of kubernetes workloads.
type KindInterface interface {
	Get(dataMap string) (string, error)
}

// ImagesInterface implements method that gets images from various kubernetes workloads.
type ImagesInterface interface {
	Get(dataMap string) (*Image, error)
}

// Image holds information of images retrieved.
type Image struct {
	Kind  string   `json:"kind,omitempty" yaml:"kind,omitempty"`
	Name  string   `json:"name,omitempty" yaml:"name,omitempty"`
	Image []string `json:"image,omitempty" yaml:"image,omitempty"`
}

func (kin *Kind) Get(dataMap string) (string, error) {
	var kindYaml map[string]interface{}

	if err := yaml.Unmarshal([]byte(dataMap), &kindYaml); err != nil {
		return "", err
	}

	if len(kindYaml) != 0 {
		value, ok := kindYaml[kubeKind].(string)
		if !ok {
			return "", &imgErrors.ImageError{Message: "failed to get name from the manifest, 'kind' is not type string"}
		}

		return value, nil
	}

	return "", nil
}

// Get identifies images from Deployments.
func (dep *Deployments) Get(dataMap string) (*Image, error) {
	if err := yaml.Unmarshal([]byte(dataMap), &dep); err != nil {
		return nil, err
	}

	depContainers := containers{append(dep.Spec.Template.Spec.Containers, dep.Spec.Template.Spec.InitContainers...)}

	images := &Image{
		Kind:  KindDeployment,
		Name:  dep.Name,
		Image: depContainers.getImages(),
	}

	return images, nil
}

// Get identifies images from StatefulSets.
func (dep *StatefulSets) Get(dataMap string) (*Image, error) {
	if err := yaml.Unmarshal([]byte(dataMap), &dep); err != nil {
		return nil, err
	}

	depContainers := containers{append(dep.Spec.Template.Spec.Containers, dep.Spec.Template.Spec.InitContainers...)}

	images := &Image{
		Kind:  KindStatefulSet,
		Name:  dep.Name,
		Image: depContainers.getImages(),
	}

	return images, nil
}

// Get identifies images from DaemonSets.
func (dep *DaemonSets) Get(dataMap string) (*Image, error) {
	if err := yaml.Unmarshal([]byte(dataMap), &dep); err != nil {
		return nil, err
	}

	depContainers := containers{append(dep.Spec.Template.Spec.Containers, dep.Spec.Template.Spec.InitContainers...)}

	images := &Image{
		Kind:  KindDaemonSet,
		Name:  dep.Name,
		Image: depContainers.getImages(),
	}

	return images, nil
}

// Get identifies images from CronJob.
func (dep *CronJob) Get(dataMap string) (*Image, error) {
	if err := yaml.Unmarshal([]byte(dataMap), &dep); err != nil {
		return nil, err
	}

	depContainers := containers{append(dep.Spec.JobTemplate.Spec.Template.Spec.Containers,
		dep.Spec.JobTemplate.Spec.Template.Spec.InitContainers...)}

	images := &Image{
		Kind:  KindCronJob,
		Name:  dep.Name,
		Image: depContainers.getImages(),
	}

	return images, nil
}

// Get identifies images from Job.
func (dep *Job) Get(dataMap string) (*Image, error) {
	if err := yaml.Unmarshal([]byte(dataMap), &dep); err != nil {
		return nil, err
	}

	depContainers := containers{append(dep.Spec.Template.Spec.Containers, dep.Spec.Template.Spec.InitContainers...)}

	images := &Image{
		Kind:  KindJob,
		Name:  dep.Name,
		Image: depContainers.getImages(),
	}

	return images, nil
}

// Get identifies images from ReplicaSets.
func (dep *ReplicaSets) Get(dataMap string) (*Image, error) {
	if err := yaml.Unmarshal([]byte(dataMap), &dep); err != nil {
		return nil, err
	}

	depContainers := containers{append(dep.Spec.Template.Spec.Containers, dep.Spec.Template.Spec.InitContainers...)}

	images := &Image{
		Kind:  KindReplicaSet,
		Name:  dep.Name,
		Image: depContainers.getImages(),
	}

	return images, nil
}

// Get identifies images from Pod.
func (dep *Pod) Get(dataMap string) (*Image, error) {
	if err := yaml.Unmarshal([]byte(dataMap), &dep); err != nil {
		return nil, err
	}

	depContainers := containers{append(dep.Spec.Containers, dep.Spec.InitContainers...)}

	images := &Image{
		Kind:  KindPod,
		Name:  dep.Name,
		Image: depContainers.getImages(),
	}

	return images, nil
}

// Get identifies images from AlertManager.
func (dep *AlertManager) Get(dataMap string) (*Image, error) {
	if err := yaml.Unmarshal([]byte(dataMap), &dep); err != nil {
		return nil, err
	}

	images := &Image{
		Kind:  monitoringV1.AlertmanagersKind,
		Name:  dep.Name,
		Image: []string{*dep.Spec.Image},
	}

	return images, nil
}

// Get identifies images from Prometheus.
func (dep *Prometheus) Get(dataMap string) (*Image, error) {
	if err := yaml.Unmarshal([]byte(dataMap), &dep); err != nil {
		return nil, err
	}

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
func (dep *ThanosRuler) Get(dataMap string) (*Image, error) {
	if err := yaml.Unmarshal([]byte(dataMap), &dep); err != nil {
		return nil, err
	}

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
func (dep *Grafana) Get(dataMap string) (*Image, error) {
	if err := yaml.Unmarshal([]byte(dataMap), &dep); err != nil {
		return nil, err
	}

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
func (dep *Thanos) Get(dataMap string) (*Image, error) {
	if err := yaml.Unmarshal([]byte(dataMap), &dep); err != nil {
		return nil, err
	}

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
func (dep *ThanosReceiver) Get(dataMap string) (*Image, error) {
	if err := yaml.Unmarshal([]byte(dataMap), &dep); err != nil {
		return nil, err
	}

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

// NewKind returns new instance of Kind.
func NewKind() KindInterface {
	return &Kind{}
}

func SupportedKinds() []string {
	kinds := []string{
		KindDeployment, KindStatefulSet, KindDaemonSet,
		KindCronJob, KindJob, KindReplicaSet, KindPod,
		monitoringV1.AlertmanagersKind, monitoringV1.PrometheusesKind, monitoringV1.ThanosRulerKind,
		KindGrafana, KindThanos, KindThanosReceiver,
	}

	return kinds
}

func (cont containers) getImages() []string {
	images := make([]string, 0)
	for _, container := range cont.containers {
		images = append(images, container.Image)
	}

	return images
}
