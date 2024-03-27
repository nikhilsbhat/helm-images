package pkg_test

import (
	"testing"

	"github.com/sboutet06/helm-images/pkg"
	"github.com/sboutet06/helm-images/pkg/k8s"
	"github.com/stretchr/testify/assert"
)

func TestImages_filterImagesByRegistries(t *testing.T) {
	t.Run("should be able to return the filtered images by registries", func(t *testing.T) {
		imageKind := k8s.Image{
			Kind: "Deployment",
			Name: "sample-deployment",
			Image: []string{
				"quay.io/prometheus/node-exporter:v1.1.2",
				"k8s.gcr.io/kube-state-metrics/kube-state-metrics:v2.0.0",
				"quay.io/prometheus/alertmanager:v0.21.0",
				"prom/pushgateway:v1.3.1",
				"jimmidyson/configmap-reload:v0.5.0",
			},
		}

		imageList := []*k8s.Image{&imageKind}

		imageClient := pkg.Images{
			Registries: []string{"quay.io", "k8s.gcr.io"},
		}
		imageClient.SetLogger("info")

		expectedImageKind := k8s.Image{
			Kind: "Deployment",
			Name: "sample-deployment",
			Image: []string{
				"quay.io/prometheus/node-exporter:v1.1.2",
				"k8s.gcr.io/kube-state-metrics/kube-state-metrics:v2.0.0",
				"quay.io/prometheus/alertmanager:v0.21.0",
			},
		}

		expected := []*k8s.Image{&expectedImageKind}

		imagesFiltered := imageClient.FilterImagesByRegistries(imageList)
		assert.ElementsMatch(t, expected[0].Image, imagesFiltered[0].Image)
	})

	t.Run("should be able to return the filtered images by registries", func(t *testing.T) {
		imageKind := k8s.Image{
			Kind: "Deployment",
			Name: "sample-deployment",
			Image: []string{
				"quay.io/prometheus/node-exporter:v1.1.2",
				"k8s.gcr.io/kube-state-metrics/kube-state-metrics:v2.0.0",
				"quay.io/prometheus/alertmanager:v0.21.0",
				"prom/pushgateway:v1.3.1",
				"jimmidyson/configmap-reload:v0.5.0",
			},
		}

		imageList := []*k8s.Image{&imageKind}

		imageClient := pkg.Images{
			Registries: []string{"qquay.io"},
		}

		imageClient.SetLogger("info")

		imagesFiltered := imageClient.FilterImagesByRegistries(imageList)
		assert.Nil(t, imagesFiltered)
	})

	t.Run("should be able to return the filtered unique images by registries", func(t *testing.T) {
		imageKind := k8s.Image{
			Kind: "Deployment",
			Name: "sample-deployment",
			Image: []string{
				"quay.io/prometheus/node-exporter:v1.1.2",
				"quay.io/prometheus/node-exporter:v1.1.2",
				"k8s.gcr.io/kube-state-metrics/kube-state-metrics:v2.0.0",
				"quay.io/prometheus/alertmanager:v0.21.0",
				"prom/pushgateway:v1.3.1",
				"jimmidyson/configmap-reload:v0.5.0",
			},
		}

		imageList := []*k8s.Image{&imageKind}

		imageClient := pkg.Images{
			Registries:   []string{"quay.io"},
			UniqueImages: true,
		}

		imageClient.SetLogger("info")

		expectedImageKind := k8s.Image{
			Kind: "Deployment",
			Name: "sample-deployment",
			Image: []string{
				"quay.io/prometheus/node-exporter:v1.1.2",
				"quay.io/prometheus/alertmanager:v0.21.0",
			},
		}
		expected := []*k8s.Image{&expectedImageKind}

		imagesFiltered := imageClient.FilterImagesByRegistries(imageList)
		assert.Equal(t, expected, imagesFiltered)
	})

	t.Run("should be image list as it is as no registries matched", func(t *testing.T) {
		imageKind := k8s.Image{
			Kind: "Deployment",
			Name: "sample-deployment",
			Image: []string{
				"quay.io/prometheus/node-exporter:v1.1.2",
				"k8s.gcr.io/kube-state-metrics/kube-state-metrics:v2.0.0",
				"quay.io/prometheus/alertmanager:v0.21.0",
				"prom/pushgateway:v1.3.1",
				"jimmidyson/configmap-reload:v0.5.0",
			},
		}

		imageList := []*k8s.Image{&imageKind}

		imageClient := pkg.Images{}

		imageClient.SetLogger("info")

		imagesFiltered := imageClient.FilterImagesByRegistries(imageList)
		assert.ElementsMatch(t, imageList[0].Image, imagesFiltered[0].Image)
	})

	t.Run("should be able to return the unique and not filtered images by registries", func(t *testing.T) {
		imageKind := k8s.Image{
			Kind: "Deployment",
			Name: "sample-deployment",
			Image: []string{
				"quay.io/prometheus/node-exporter:v1.1.2",
				"quay.io/prometheus/node-exporter:v1.1.2",
				"k8s.gcr.io/kube-state-metrics/kube-state-metrics:v2.0.0",
				"quay.io/prometheus/alertmanager:v0.21.0",
				"prom/pushgateway:v1.3.1",
				"jimmidyson/configmap-reload:v0.5.0",
			},
		}

		imageList := []*k8s.Image{&imageKind}

		imageClient := pkg.Images{
			UniqueImages: true,
		}
		imageClient.SetLogger("info")

		expectedImageKind := k8s.Image{
			Kind: "Deployment",
			Name: "sample-deployment",
			Image: []string{
				"quay.io/prometheus/node-exporter:v1.1.2",
				"k8s.gcr.io/kube-state-metrics/kube-state-metrics:v2.0.0",
				"quay.io/prometheus/alertmanager:v0.21.0",
				"prom/pushgateway:v1.3.1",
				"jimmidyson/configmap-reload:v0.5.0",
			},
		}

		expected := []*k8s.Image{&expectedImageKind}

		imagesFiltered := imageClient.FilterImagesByRegistries(imageList)
		assert.ElementsMatch(t, expected[0].Image, imagesFiltered[0].Image)
	})
}

func Test_filteredImages(t *testing.T) {
	t.Run("should be able to filter the images by the list of registries", func(t *testing.T) {
		images := []string{
			"quay.io/prometheus/node-exporter:v1.1.2",
			"k8s.gcr.io/kube-state-metrics/kube-state-metrics:v2.0.0",
			"quay.io/prometheus/alertmanager:v0.21.0",
		}
		registries := []string{"quay.io"}

		expected := []string{
			"quay.io/prometheus/alertmanager:v0.21.0",
			"quay.io/prometheus/node-exporter:v1.1.2",
		}

		actual := pkg.FilteredImages(images, registries)
		assert.ElementsMatch(t, actual, expected)
	})

	t.Run("should be able to filter the images by the list of registries with matching names", func(t *testing.T) {
		images := []string{
			"quay.io/prometheus/node-exporter:v1.1.2",
			"k8s.gcr.io/kube-state-metrics/kube-state-metrics:v2.0.0",
			"quay.io/prometheus/alertmanager:v0.21.0",
		}
		registries := []string{"quay"}

		expected := []string{
			"quay.io/prometheus/alertmanager:v0.21.0",
			"quay.io/prometheus/node-exporter:v1.1.2",
		}

		actual := pkg.FilteredImages(images, registries)
		assert.ElementsMatch(t, actual, expected)
	})
}
