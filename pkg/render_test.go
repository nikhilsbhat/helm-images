package pkg

import (
	"bytes"
	"testing"

	"github.com/nikhilsbhat/helm-images/pkg/k8s"
	"github.com/stretchr/testify/assert"
)

func TestImages_toYAML(t *testing.T) {
	t.Run("should be able to render the output in yaml format", func(t *testing.T) {
		yamlOut := &bytes.Buffer{}
		imageClient := Images{}
		imageClient.SetWriter(yamlOut)
		imageClient.SetLogger("info")

		images := k8s.Image{
			Kind: "Deployment",
			Name: "sample-deployment",
			Image: []string{
				"prom/pushgateway:v1.3.1",
				"jimmidyson/configmap-reload:v0.5.0",
			},
		}
		imagesFiltered := []*k8s.Image{&images}

		expected := "---\n- image:\n  - prom/pushgateway:v1.3.1\n  - jimmidyson/configmap-reload:v0.5.0\n  kind: Deployment\n  name: sample-deployment\n"
		err := imageClient.toYAML(imagesFiltered)
		assert.NoError(t, err)
		assert.Equal(t, expected, yamlOut.String())
	})
}

func TestImages_toJSON(t *testing.T) {
	t.Run("should be able to render the output in json format", func(t *testing.T) {
		jsonOut := &bytes.Buffer{}
		imageClient := Images{}
		imageClient.SetWriter(jsonOut)
		imageClient.SetLogger("info")

		images := k8s.Image{
			Kind: "Deployment",
			Name: "sample-deployment",
			Image: []string{
				"prom/pushgateway:v1.3.1",
				"jimmidyson/configmap-reload:v0.5.0",
			},
		}
		imagesFiltered := []*k8s.Image{&images}

		expected := "[\n  {\n   \"kind\": \"Deployment\",\n   \"name\": \"sample-deployment\",\n   \"image\": [\n    \"prom/pushgateway:v1.3.1\",\n    \"jimmidyson/configmap-reload:v0.5.0\"\n   ]\n  }\n ]"
		err := imageClient.toJSON(imagesFiltered)
		assert.NoError(t, err)
		assert.Equal(t, expected, jsonOut.String())
	})
}
