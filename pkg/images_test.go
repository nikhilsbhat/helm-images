package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImages_filterImages(t *testing.T) {
	t.Run("should be able to return the filtered image list", func(t *testing.T) {
		imageList := []string{
			"quay.io/prometheus/node-exporter:v1.1.2",
			"quay.io/prometheus/alertmanager:v0.21.0",
			"k8s.gcr.io/kube-state-metrics/kube-state-metrics:v2.0.0",
			"prom/pushgateway:v1.3.1",
			"jimmidyson/configmap-reload:v0.5.0",
		}

		imageClient := Images{
			Registries: []string{"quay.io", "k8s.gcr.io"},
		}

		expected := []string{
			"quay.io/prometheus/node-exporter:v1.1.2",
			"quay.io/prometheus/alertmanager:v0.21.0",
			"k8s.gcr.io/kube-state-metrics/kube-state-metrics:v2.0.0",
		}

		filteredImages := imageClient.filterImages(imageList)
		assert.Equal(t, expected, filteredImages)
	})
}

func Test_getImages(t *testing.T) {
	helmTemplate := `---
# Source: prometheus/charts/prometheus/templates/alertmanager/clusterrole.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    component: "alertmanager"
    app: prometheus
    release: prometheus-standalone
    chart: prometheus-14.4.1
    heritage: Helm
  name: prometheus-standalone-alertmanager
rules:
  []
---
# Source: prometheus/charts/prometheus/templates/pushgateway/clusterrole.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    component: "pushgateway"
    app: prometheus
    release: prometheus-standalone
    chart: prometheus-14.4.1
    heritage: Helm
  name: prometheus-standalone-pushgateway
rules:
  []
---
# Source: prometheus/charts/prometheus/charts/kube-state-metrics/templates/clusterrolebinding.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  labels:
    app.kubernetes.io/name: kube-state-metrics
    helm.sh/chart: kube-state-metrics-3.1.1
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/instance: prometheus-standalone
  name: prometheus-standalone-kube-state-metrics
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: prometheus-standalone-kube-state-metrics
subjects:
- kind: ServiceAccount
  name: prometheus-standalone-kube-state-metrics
  namespace: test`
	t.Run("", func(t *testing.T) {
		expected := []string{
			"\n# Source: prometheus/charts/prometheus/templates/alertmanager/clusterrole.yaml\napiVersion: rbac.authorization.k8s.io/v1\nkind: ClusterRole\nmetadata:\n  labels:\n    component: \"alertmanager\"\n    app: prometheus\n    release: prometheus-standalone\n    chart: prometheus-14.4.1\n    heritage: Helm\n  name: prometheus-standalone-alertmanager\nrules:\n  []\n",                                                                                                                                                                                                                                                                                                  //nolint:lll
			"\n# Source: prometheus/charts/prometheus/templates/pushgateway/clusterrole.yaml\napiVersion: rbac.authorization.k8s.io/v1\nkind: ClusterRole\nmetadata:\n  labels:\n    component: \"pushgateway\"\n    app: prometheus\n    release: prometheus-standalone\n    chart: prometheus-14.4.1\n    heritage: Helm\n  name: prometheus-standalone-pushgateway\nrules:\n  []\n",                                                                                                                                                                                                                                                                                                     //nolint:lll
			"\n# Source: prometheus/charts/prometheus/charts/kube-state-metrics/templates/clusterrolebinding.yaml\napiVersion: rbac.authorization.k8s.io/v1\nkind: ClusterRoleBinding\nmetadata:\n  labels:\n    app.kubernetes.io/name: kube-state-metrics\n    helm.sh/chart: kube-state-metrics-3.1.1\n    app.kubernetes.io/managed-by: Helm\n    app.kubernetes.io/instance: prometheus-standalone\n  name: prometheus-standalone-kube-state-metrics\nroleRef:\n  apiGroup: rbac.authorization.k8s.io\n  kind: ClusterRole\n  name: prometheus-standalone-kube-state-metrics\nsubjects:\n- kind: ServiceAccount\n  name: prometheus-standalone-kube-state-metrics\n  namespace: test", //nolint:lll
		}
		actual := getTemplates([]byte(helmTemplate))
		assert.ElementsMatch(t, expected, actual)
	})
}

func TestImages_getReleaseNChart(t *testing.T) {
	t.Run("should be able to fetch release and chart from the arguments passed", func(t *testing.T) {
		imageClient := Images{
			Registries: []string{"quay.io", "k8s.gcr.io"},
		}

		expected := Images{
			Registries: []string{"quay.io", "k8s.gcr.io"},
			release:    "test-chart",
			chart:      "path/to/chart",
		}

		err := imageClient.getReleaseNChart([]string{"test-chart", "path/to/chart"})
		assert.Nil(t, err)
		assert.Equal(t, expected, imageClient)
	})
	t.Run("should error out with missing argument", func(t *testing.T) {
		imageClient := Images{
			Registries: []string{"quay.io", "k8s.gcr.io"},
		}

		err := imageClient.getReleaseNChart([]string{"test-chart"})
		assert.EqualError(t, err, "[RELEASE] or [CHART] cannot be empty")
	})
}
