package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_find(t *testing.T) {
	t.Run("should be able to find element in the slice", func(t *testing.T) {
		slice := []string{"first", "second", "third"}

		actual := find(slice, "first")
		assert.Equal(t, true, actual)
	})
	t.Run("should return false as element is missing", func(t *testing.T) {
		slice := []string{"first", "second", "third"}

		actual := find(slice, "fourth")
		assert.Equal(t, false, actual)
	})
}

func Test_findKey(t *testing.T) {
	t.Run("should be able to find the value for the key entered", func(t *testing.T) {
		sampleMap := map[string]interface{}{
			"type":       "Opaque",
			"kind":       "Secret",
			"apiVersion": "v1",
			"data": map[string]interface{}{
				"admin-password": "T0VWSnIxQWFscklYYjlNczJHcWZ3ZjRDUERyY2V3U3RUaE51RklJYg==",
				"admin-user":     "YWRtaW4==",
			},
			"spec": map[string]interface{}{
				"command":         "/opt/bats/bin/bats -t /tests/run.sh",
				"image":           "bats/bats:v1.1.0",
				"imagePullPolicy": "IfNotPresent",
				"name":            "/Users/nikhil.bhat/grafana-helm-test",
				"volumeMounts": map[string]interface{}{
					"mountPath": "/tests",
					"name":      "tests",
					"readOnly":  true,
				},
			},
		}

		expected := "bats/bats:v1.1.0"
		actual, status := findKey(sampleMap, "image")
		assert.Equal(t, true, status)
		assert.Equal(t, expected, actual)
	})
	t.Run("should be able to find the value for the key entered", func(t *testing.T) {
		sampleMap := map[string]interface{}{
			"type":       "Opaque",
			"kind":       "Secret",
			"apiVersion": "v1",
			"data": map[string]interface{}{
				"admin-password": "T0VWSnIxQWFscklYYjlNczJHcWZ3ZjRDUERyY2V3U3RUaE51RklJYg==",
				"admin-user":     "YWRtaW4==",
			},
			"spec": map[string]interface{}{
				"command":         "/opt/bats/bin/bats -t /tests/run.sh",
				"imagePullPolicy": "IfNotPresent",
				"name":            "/Users/nikhil.bhat/grafana-helm-test",
				"volumeMounts": map[string]interface{}{
					"mountPath": "/tests",
					"name":      "tests",
					"readOnly":  true,
				},
			},
		}

		actual, status := findKey(sampleMap, "image")
		assert.Equal(t, false, status)
		assert.Nil(t, actual)
	})
}

func Test_getUniqueSlice(t *testing.T) {
	t.Run("should remove duplicates from the list", func(t *testing.T) {
		slice := []string{"one", "two", "three", "three", "one", "four"}
		expected := []string{"one", "two", "three", "four"}
		actual := getUniqueSlice(slice)
		assert.ElementsMatch(t, expected, actual)
	})
}
