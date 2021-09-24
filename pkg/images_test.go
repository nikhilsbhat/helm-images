package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImages_filterImages(t *testing.T) {
	t.Run("should be able to return the filtered image list", func(t *testing.T) {
		imageList := []kind{
			{
				Image: "quay.io/prometheus/node-exporter:v1.1.2",
			},
			{
				Image: "k8s.gcr.io/kube-state-metrics/kube-state-metrics:v2.0.0",
			},
			{
				Image: "quay.io/prometheus/alertmanager:v0.21.0",
			},
			{
				Image: "prom/pushgateway:v1.3.1",
			},
			{
				Image: "jimmidyson/configmap-reload:v0.5.0",
			},
		}

		imageClient := Images{
			Registries: []string{"quay.io", "k8s.gcr.io"},
		}

		expected := []kind{
			{
				Image: "quay.io/prometheus/node-exporter:v1.1.2",
			},
			{
				Image: "k8s.gcr.io/kube-state-metrics/kube-state-metrics:v2.0.0",
			},
			{
				Image: "quay.io/prometheus/alertmanager:v0.21.0",
			},
		}
		filteredImages := imageClient.filterImages(imageList)
		assert.ElementsMatch(t, expected, filteredImages)
	})
}

func Test_getImages(t *testing.T) {
	imageClient := Images{
		ImageRegex: ImageRegex,
	}
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
		expected := []string{
			"\napiVersion: rbac.authorization.k8s.io/v1\nkind: ClusterRole\nmetadata:\n  labels:\n    component: \"alertmanager\"\n    app: prometheus\n    release: prometheus-standalone\n    chart: prometheus-14.4.1\n    heritage: Helm\n  name: prometheus-standalone-alertmanager\nrules:\n  []\n",                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    //nolint:lll
			"\napiVersion: rbac.authorization.k8s.io/v1\nkind: ClusterRole\nmetadata:\n  labels:\n    component: \"pushgateway\"\n    app: prometheus\n    release: prometheus-standalone\n    chart: prometheus-14.4.1\n    heritage: Helm\n  name: prometheus-standalone-pushgateway\nrules:\n  []\n",                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      //nolint:lll
			"\napiVersion: rbac.authorization.k8s.io/v1\nkind: ClusterRoleBinding\nmetadata:\n  labels:\n    app.kubernetes.io/name: kube-state-metrics\n    helm.sh/chart: kube-state-metrics-3.1.1\n    app.kubernetes.io/managed-by: Helm\n    app.kubernetes.io/instance: prometheus-standalone\n  name: prometheus-standalone-kube-state-metrics\nroleRef:\n  apiGroup: rbac.authorization.k8s.io\n  kind: ClusterRole\n  name: prometheus-standalone-kube-state-metrics\nsubjects:\n- kind: ServiceAccount\n  name: prometheus-standalone-kube-state-metrics\n  namespace: test\n",                                                                                                                                                                                                                                                                                                                     //nolint:lll
			"\napiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: jaeger-ca-cert\ndata:\n    CA_CERTIFICATE: |\n        -----BEGIN CERTIFICATE-----\n\t\tOCOIRRGVEGHEIGHEnwoircne20394809234nfh834retitneh83t5ljfKHD&$&$\n\t\tOCOIRRGVEGHEIGHEnwoircne20394809234nfh834retitneh83t5ljfKHD&$&$\n\t\tOCOIRRGVEGHEIGHEnwoircne20394809234nfh834retitneh83t5ljfKHD&$&$\n\t\tOCOIRRGVEGHEIGHEnwoircne20394809234nfh834retitneh83t5ljfKHD&$&$\n\t\tOCOIRRGVEGHEIGHEnwoircne20394809234nfh834retitneh83t5ljfKHD&$&$\n\t\tOCOIRRGVEGHEIGHEnwoircne20394809234nfh834retitneh83t5ljfKHD&$&$\n\t\tOCOIRRGVEGHEIGHEnwoircne20394809234nfh834retitneh83t5ljfKHD&$&$\n\t\tOCOIRRGVEGHEIGHEnwoircne20394809234nfh834retitneh83t5ljfKHD&$&$\n\t\tOCOIRRGVEGHEIGHEnwoircne20394809234nfh834retitneh83t5ljfKHD&$&$\n\t\tOCOIRRGVEGHEIGHEnwoircne20394809234nfh834retitneh83t5ljfKHD&$&$\n        -----END CERTIFICATE-----\n", //nolint:lll
		}
		actual := imageClient.getTemplates([]byte(helmTemplate))
		assert.ElementsMatch(t, expected, actual)
	})
}

func Test_getImagesFromKind(t *testing.T) {
	t.Run("should be able to get all images from struct kind", func(t *testing.T) {
		kindObj := []kind{
			{Kind: "DaemonSet", Name: "prometheus-standalone-node-exporter", Image: "quay.io/prometheus/node-exporter:v1.1.2"},
			{Kind: "Deployment", Name: "prometheus-standalone-server", Image: "jimmidyson/configmap-reload:v0.5.0"},
			{Kind: "StatefulSet", Name: "prometheus-standalone-kube-state-metrics", Image: "k8s.gcr.io/kube-state-metrics/kube-state-metrics:v2.0.0"},
		}

		expected := []string{"quay.io/prometheus/node-exporter:v1.1.2", "jimmidyson/configmap-reload:v0.5.0", "k8s.gcr.io/kube-state-metrics/kube-state-metrics:v2.0.0"}
		images := getImagesFromKind(kindObj)
		assert.ElementsMatch(t, expected, images)
	})
}
