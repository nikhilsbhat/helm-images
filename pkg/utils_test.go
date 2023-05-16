package pkg_test

import (
	"testing"

	"github.com/nikhilsbhat/helm-images/pkg"
	"github.com/stretchr/testify/assert"
)

func Test_getUniqEntries(t *testing.T) {
	t.Run("should filter struct slice to get unique entries", func(t *testing.T) {
		sampleMap := []string{
			"quay.io/prometheus/node-exporter:v1.1.2",
			"k8s.gcr.io/kube-state-metrics/kube-state-metrics:v2.0.0",
			"quay.io/prometheus/alertmanager:v0.21.0",
			"quay.io/prometheus/alertmanager:v0.21.0",
			"quay.io/prometheus/alertmanager:v0.21.0",
		}

		expected := []string{
			"quay.io/prometheus/node-exporter:v1.1.2",
			"k8s.gcr.io/kube-state-metrics/kube-state-metrics:v2.0.0",
			"quay.io/prometheus/alertmanager:v0.21.0",
		}
		actual := pkg.GetUniqEntries(sampleMap)
		assert.Equal(t, expected, actual)
	})
}

func Test_contains(t *testing.T) {
	t.Run("should return true as struct Contains the element", func(t *testing.T) {
		sampleMap := []string{
			"quay.io/prometheus/node-exporter:v1.1.2",
			"k8s.gcr.io/kube-state-metrics/kube-state-metrics:v2.0.0",
			"quay.io/prometheus/alertmanager:v0.21.0",
			"quay.io/prometheus/alertmanager:v0.21.0",
		}
		actual := pkg.Contains(sampleMap, "quay.io/prometheus/alertmanager:v0.21.0")
		assert.Equal(t, true, actual)
	})
}
