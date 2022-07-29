package k8s

import (
	"github.com/ghodss/yaml"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
)

const (
	KindDeployment  = "Deployment"
	KindStatefulSet = "StatefulSet"
	KindDaemonSet   = "DaemonSet"
	KindCronJob     = "CronJob"
	KindJob         = "Job"
	KindReplicaSet  = "ReplicaSet"
	kubeKind        = "kind"
)

type (
	Deployments  appsv1.Deployment
	StatefulSets appsv1.StatefulSet
	DaemonSets   appsv1.DaemonSet
	ReplicaSets  appsv1.ReplicaSet
	CronJob      batchv1.CronJob
	Job          batchv1.Job
	Kind         map[string]interface{}
	containers   struct {
		containers []v1.Container
	}
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
		return kindYaml[kubeKind].(string), nil
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
