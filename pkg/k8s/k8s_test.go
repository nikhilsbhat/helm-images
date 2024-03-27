package k8s_test

import (
	"encoding/json"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/sboutet06/helm-images/pkg/k8s"
	"github.com/stretchr/testify/assert"
)

func TestGetVal(t *testing.T) {
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
		assert.NoError(t, err)
		valueFound, _ := k8s.GetImage(valueMap, "image")
		assert.ElementsMatch(t, []string{
			"ghcr.io/example/config:v2.3.0",
			"ghcr.io/example/testConfig:v2.3.0",
			"ghcr.io/example/sample:v2.2.0",
		}, valueFound)
	})

	t.Run("should be able to fetch the image from json string", func(t *testing.T) {
		valueMap := make(map[string]interface{})

		err := json.Unmarshal([]byte(jsonData), &valueMap)
		assert.NoError(t, err)
		valueFound, _ := k8s.GetImage(valueMap, "image")
		assert.ElementsMatch(t, []string{
			"ghcr.io/prometheus/prom:v2.0.0",
			"ghcr.io/example/sample:v2.2.0",
			"ghcr.io/example/config:v2.3.0",
		}, valueFound)
	})
}
