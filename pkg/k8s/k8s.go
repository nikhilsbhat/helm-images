package k8s

import (
	"github.com/ghodss/yaml"
	imgErrors "github.com/nikhilsbhat/helm-images/pkg/errors"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	KindDeployment  = "Deployment"
	KindStatefulSet = "StatefulSet"
	KindDaemonSet   = "DaemonSet"
	KindCronJob     = "CronJob"
	KindJob         = "Job"
	KindReplicaSet  = "ReplicaSet"
	KindPod         = "Pod"
	kubeKind        = "kind"
)

type (
	Deployments  appsv1.Deployment
	StatefulSets appsv1.StatefulSet
	DaemonSets   appsv1.DaemonSet
	ReplicaSets  appsv1.ReplicaSet
	CronJob      batchv1.CronJob
	Job          batchv1.Job
	Pod          corev1.Pod
	Kind         map[string]interface{}
	containers   struct {
		containers []corev1.Container
	}
	AlertManager monitoringv1.Alertmanager
	Prometheus   monitoringv1.Prometheus
	ThanosRuler  monitoringv1.ThanosRuler
)

type KindInterface interface {
	Get(dataMap string) (string, error)
}

type ImagesInterface interface {
	Get(dataMap string) (*Image, error)
}

type Image struct {
	Kind  string   `json:"kind,omitempty"`
	Name  string   `json:"name,omitempty"`
	Image []string `json:"image,omitempty"`
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

func (dep *AlertManager) Get(dataMap string) (*Image, error) {
	if err := yaml.Unmarshal([]byte(dataMap), &dep); err != nil {
		return nil, err
	}

	images := &Image{
		Kind:  monitoringv1.AlertmanagersKind,
		Name:  dep.Name,
		Image: []string{*dep.Spec.Image},
	}

	return images, nil
}

func (dep *Prometheus) Get(dataMap string) (*Image, error) {
	if err := yaml.Unmarshal([]byte(dataMap), &dep); err != nil {
		return nil, err
	}

	var imageNames []string

	depContainers := containers{append(dep.Spec.Containers, dep.Spec.InitContainers...)}

	imageNames = append(imageNames, depContainers.getImages()...)
	imageNames = append(imageNames, *dep.Spec.Image)

	images := &Image{
		Kind:  monitoringv1.PrometheusesKind,
		Name:  dep.Name,
		Image: imageNames,
	}

	return images, nil
}

func (dep *ThanosRuler) Get(dataMap string) (*Image, error) {
	if err := yaml.Unmarshal([]byte(dataMap), &dep); err != nil {
		return nil, err
	}

	var imageNames []string

	depContainers := containers{append(dep.Spec.Containers, dep.Spec.InitContainers...)}

	imageNames = append(imageNames, depContainers.getImages()...)
	imageNames = append(imageNames, dep.Spec.Image)

	images := &Image{
		Kind:  monitoringv1.ThanosRulerKind,
		Name:  dep.Name,
		Image: imageNames,
	}

	return images, nil
}

func NewDeployment() ImagesInterface {
	return &Deployments{}
}

func NewStatefulSet() ImagesInterface {
	return &StatefulSets{}
}

func NewDaemonSet() ImagesInterface {
	return &DaemonSets{}
}

func NewReplicaSets() ImagesInterface {
	return &ReplicaSets{}
}

func NewCronjob() ImagesInterface {
	return &CronJob{}
}

func NewJob() ImagesInterface {
	return &Job{}
}

func NewPod() ImagesInterface {
	return &Pod{}
}

func NewAlertManager() ImagesInterface {
	return &AlertManager{}
}

func NewPrometheus() ImagesInterface {
	return &Prometheus{}
}

func NewThanosRuler() ImagesInterface {
	return &ThanosRuler{}
}

func NewKind() KindInterface {
	return &Kind{}
}

func (cont containers) getImages() []string {
	images := make([]string, 0)
	for _, container := range cont.containers {
		images = append(images, container.Image)
	}

	return images
}
