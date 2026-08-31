package k8s_test

import (
	"encoding/json"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/nikhilsbhat/helm-images/pkg"
	"github.com/nikhilsbhat/helm-images/pkg/k8s"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetVal(t *testing.T) {
	log := logrus.New()
	yamlData := `image: 'ghcr.io/example/sample:v2.2.0'
enemies: aliens
lives: '3'
config:
  image: 'ghcr.io/example/config:v2.3.0'
  testConfig:
    image: 'ghcr.io/example/testConfig:v2.3.0'`

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
			"ghcr.io/example/testConfig:v2.3.0",
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
