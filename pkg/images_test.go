package pkg_test

import (
	"testing"

	"github.com/nikhilsbhat/helm-images/pkg"
	"github.com/nikhilsbhat/helm-images/pkg/k8s"
	"github.com/stretchr/testify/assert"
)

func Test_getImages(t *testing.T) {
	imageClient := pkg.Images{
		ImageRegex: pkg.ImageRegex,
	}
	imageClient.SetLogger("info")

	helmTemplate := `
---
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
  namespace: test
---
# Empty template
# Just comments
---
# Source: tracing/templates/jaeger/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
 name: jaeger-ca-cert
data:
   CA_CERTIFICATE: |
       -----BEGIN CERTIFICATE-----
       OCOIRRGVEGHEIGHEnwoircne20394809234nfh834retitneh83t5ljfKHD&$&$
       OCOIRRGVEGHEIGHEnwoircne20394809234nfh834retitneh83t5ljfKHD&$&$
       OCOIRRGVEGHEIGHEnwoircne20394809234nfh834retitneh83t5ljfKHD&$&$
       OCOIRRGVEGHEIGHEnwoircne20394809234nfh834retitneh83t5ljfKHD&$&$
       OCOIRRGVEGHEIGHEnwoircne20394809234nfh834retitneh83t5ljfKHD&$&$
       OCOIRRGVEGHEIGHEnwoircne20394809234nfh834retitneh83t5ljfKHD&$&$
       OCOIRRGVEGHEIGHEnwoircne20394809234nfh834retitneh83t5ljfKHD&$&$
       OCOIRRGVEGHEIGHEnwoircne20394809234nfh834retitneh83t5ljfKHD&$&$
       OCOIRRGVEGHEIGHEnwoircne20394809234nfh834retitneh83t5ljfKHD&$&$
       OCOIRRGVEGHEIGHEnwoircne20394809234nfh834retitneh83t5ljfKHD&$&$
       -----END CERTIFICATE-----
`

	t.Run("should be able to split rendered templates to individual templates", func(t *testing.T) {
		actual, err := imageClient.GetTemplates([]byte(helmTemplate))
		assert.NoError(t, err)
		assert.Len(t, actual, 4)
	})
}

func Test_getImagesFromKind(t *testing.T) {
	t.Run("should be able to get all images from struct kind", func(t *testing.T) {
		kindObj := []*k8s.Image{
			{Kind: "DaemonSet", Name: "prometheus-standalone-node-exporter", Image: []string{"quay.io/prometheus/node-exporter:v1.1.2"}},
			{Kind: "Deployment", Name: "prometheus-standalone-server", Image: []string{"jimmidyson/configmap-reload:v0.5.0"}},
			{Kind: "StatefulSet", Name: "prometheus-standalone-kube-state-metrics", Image: []string{"k8s.gcr.io/kube-state-metrics/kube-state-metrics:v2.0.0"}},
		}

		expected := []string{
			"quay.io/prometheus/node-exporter:v1.1.2",
			"jimmidyson/configmap-reload:v0.5.0",
			"k8s.gcr.io/kube-state-metrics/kube-state-metrics:v2.0.0",
		}
		images := pkg.GetImagesFromKind(kindObj)
		assert.ElementsMatch(t, expected, images)
	})
}

func TestImages_SetRelease(t *testing.T) {
	t.Run("Should be able to set the release", func(t *testing.T) {
		imageClient := pkg.Images{}
		imageClient.SetRelease("testRelease")

		assert.Equal(t, "testRelease", imageClient.GetRelease())
	})
}

func TestImages_SetChart(t *testing.T) {
	t.Run("Should be able to set the chart", func(t *testing.T) {
		imageClient := pkg.Images{}
		imageClient.SetChart("testChart")

		assert.Equal(t, "testChart", imageClient.GetChart())
	})
}
