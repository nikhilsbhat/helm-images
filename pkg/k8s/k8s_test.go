package k8s_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/nikhilsbhat/helm-images/pkg"
	"github.com/nikhilsbhat/helm-images/pkg/k8s"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkloadImageVolumes(t *testing.T) {
	const podSpec = `containers:
  - name: app
    image: nginx:1.14.2
    args:
      - --acme-http01-solver-image=quay.io/example/solver:v1
initContainers:
  - name: init
    image: busybox:1.36.1
volumes:
  - name: models
    image:
      pullPolicy: IfNotPresent
      reference: my.registry/my-model:1.0.0
  - name: digest
    image:
      reference: my.registry/data@sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef
  - name: scratch
    emptyDir: {}
  - name: config
    configMap:
      name: config
  - name: empty-image
    image: {}
`
	tests := []struct {
		kind   string
		prefix string
		indent int
	}{
		{kind: k8s.KindPod, prefix: "spec:\n", indent: 2},
		{kind: k8s.KindDeployment, prefix: "spec:\n  template:\n    spec:\n", indent: 6},
		{kind: k8s.KindStatefulSet, prefix: "spec:\n  template:\n    spec:\n", indent: 6},
		{kind: k8s.KindDaemonSet, prefix: "spec:\n  template:\n    spec:\n", indent: 6},
		{kind: k8s.KindReplicaSet, prefix: "spec:\n  template:\n    spec:\n", indent: 6},
		{kind: k8s.KindJob, prefix: "spec:\n  template:\n    spec:\n", indent: 6},
		{kind: k8s.KindCronJob, prefix: "spec:\n  jobTemplate:\n    spec:\n      template:\n        spec:\n", indent: 10},
	}

	for _, test := range tests {
		t.Run(test.kind, func(t *testing.T) {
			indent := strings.Repeat(" ", test.indent)
			manifest := "kind: " + test.kind + "\nmetadata:\n  name: image-volumes\n" + test.prefix +
				indent + strings.ReplaceAll(strings.TrimSuffix(podSpec, "\n"), "\n", "\n"+indent) + "\n"

			images, err := (&pkg.Images{}).GetImage(test.kind, manifest)
			require.NoError(t, err)
			require.Len(t, images, 1)
			assert.Equal(t, test.kind, images[0].Kind)
			assert.Equal(t, "image-volumes", images[0].Name)
			assert.Equal(t, []string{
				"nginx:1.14.2",
				"busybox:1.36.1",
				"quay.io/example/solver:v1",
				"my.registry/my-model:1.0.0",
				"my.registry/data@sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			}, images[0].Image)
		})
	}
}

func TestPodWithoutImageVolumes(t *testing.T) {
	for _, volumes := range []string{"", "  volumes:\n    - name: scratch\n      emptyDir: {}\n"} {
		manifest := "kind: Pod\nmetadata:\n  name: app\nspec:\n  containers:\n    - name: app\n      image: nginx:1.14.2\n" + volumes
		images, err := k8s.NewPod().Get(manifest, "", logrus.New())
		require.NoError(t, err)
		assert.Equal(t, []string{"nginx:1.14.2"}, images.Image)
	}
}

func TestGetVal(t *testing.T) {
	log := logrus.New()
	yamlData := `image: 'ghcr.io/example/sample:v2.2.0'
enemies: aliens
lives: '3'
config:
  image: 'ghcr.io/example/config:v2.3.0'
  testConfig:
    image: 'ghcr.io/example/testconfig:v2.3.0'
documentation:
  image: 'this string mentions image but is not a container image'`

	jsonData := `{
      "prometheusImage": "ghcr.io/prometheus/prom:v2.0.0",
      "image": "ghcr.io/example/sample:v2.2.0",
      "enemies": "aliens",
      "lives": "3",
      "config": {
        "image": "ghcr.io/example/config:v2.3.0"
      }
    }`

	t.Run("should be able to fetch the image from yaml string", func(t *testing.T) {
		valueMap := make(map[string]interface{})

		err := yaml.Unmarshal([]byte(yamlData), &valueMap)
		require.NoError(t, err)

		valueFound, _ := k8s.GetImage(valueMap, "image", pkg.ConfigMapImageRegex, log)
		assert.ElementsMatch(t, []string{
			"ghcr.io/example/config:v2.3.0",
			"ghcr.io/example/testconfig:v2.3.0",
			"ghcr.io/example/sample:v2.2.0",
		}, valueFound)
	})

	t.Run("should be able to fetch the image from json string", func(t *testing.T) {
		valueMap := make(map[string]interface{})

		err := json.Unmarshal([]byte(jsonData), &valueMap)
		require.NoError(t, err)

		valueFound, _ := k8s.GetImage(valueMap, "image", pkg.ConfigMapImageRegex, log)
		assert.ElementsMatch(t, []string{
			"ghcr.io/example/sample:v2.2.0",
			"ghcr.io/example/config:v2.3.0",
		}, valueFound)
	})
}

func TestPrometheusGetWithoutThanos(t *testing.T) {
	log := logrus.New()
	prometheusManifest := `kind: Prometheus
metadata:
  name: myrelease-kube-prometheus-prometheus
spec:
  image: quay.io/prometheus/prometheus:v3.11.3
  containers:
    - name: config-reloader
      image: quay.io/prometheus-operator/prometheus-config-reloader:v0.91.0
`

	images, err := k8s.NewPrometheus().Get(prometheusManifest, "", log)
	require.NoError(t, err)
	assert.Equal(t, "myrelease-kube-prometheus-prometheus", images.Name)
	assert.ElementsMatch(t, []string{
		"quay.io/prometheus/prometheus:v3.11.3",
		"quay.io/prometheus-operator/prometheus-config-reloader:v0.91.0",
	}, images.Image)
}

func TestHelmChartConfigGet(t *testing.T) {
	log := logrus.New()
	helmChartConfigManifest := `apiVersion: helm.cattle.io/v1
kind: HelmChartConfig
metadata:
  name: traefik
  namespace: kube-system
spec:
  valuesContent: |-
    deployment:
      kind: DaemonSet
      additionalContainers:
        - name: nxlog
          image: docker.io/nxlog/nxlog-ce:3.2.2329
          imagePullPolicy: IfNotPresent
      initContainers:
        - name: setup
          image: busybox:1.36.1
    config:
      sidecar:
        image: ghcr.io/example/sidecar:v1.0.0
      note:
        image: this is a note about image settings
`

	images, err := k8s.NewHelmChartConfig().Get(helmChartConfigManifest, pkg.ConfigMapImageRegex, log)
	require.NoError(t, err)
	assert.Equal(t, k8s.KindHelmChartConfig, images.Kind)
	assert.Equal(t, "traefik", images.Name)
	assert.ElementsMatch(t, []string{
		"docker.io/nxlog/nxlog-ce:3.2.2329",
		"busybox:1.36.1",
		"ghcr.io/example/sidecar:v1.0.0",
	}, images.Image)
}

func TestHelmChartConfigGetWithoutImages(t *testing.T) {
	log := logrus.New()
	helmChartConfigManifest := `apiVersion: helm.cattle.io/v1
kind: HelmChartConfig
metadata:
  name: traefik
spec:
  valuesContent: |-
    deployment:
      kind: DaemonSet
`

	images, err := k8s.NewHelmChartConfig().Get(helmChartConfigManifest, pkg.ConfigMapImageRegex, log)
	require.NoError(t, err)
	assert.Equal(t, &k8s.Image{}, images)
}

func TestConfigMapGetSkipsInvalidImageStrings(t *testing.T) {
	log := logrus.New()
	configMapManifest := `apiVersion: v1
kind: ConfigMap
metadata:
  name: sample-config
data:
  image: ""
  image-note: this is not a container image
  image-reference: ghcr.io/example/config:v2.3.0
`

	images, err := k8s.NewConfigMap().Get(configMapManifest, pkg.ConfigMapImageRegex, log)
	require.NoError(t, err)
	assert.Equal(t, []string{"ghcr.io/example/config:v2.3.0"}, images.Image)
}
