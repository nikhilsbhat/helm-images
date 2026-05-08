package pkg

import (
	"testing"

	"github.com/nikhilsbhat/helm-images/pkg/k8s"
	"github.com/stretchr/testify/assert"
)

func TestImagesSetOutputWithPlatform(t *testing.T) {
	imageClient := Images{Platform: "linux/arm64"}
	images := []*k8s.Image{
		{
			Kind:  "Deployment",
			Name:  "sample",
			Image: []string{"docker.io/library/nginx:1.27.0"},
		},
	}

	output := imageClient.setOutput(images)

	assert.Equal(t, "docker pull --platform linux/arm64 docker.io/library/nginx:1.27.0", output)
}

func TestImagesSetOutputTableWithPlatform(t *testing.T) {
	imageClient := Images{
		OutputFormat: "table",
		Platform:     "linux/arm64",
	}
	imageClient.SetOutputFormats()
	images := []*k8s.Image{
		{
			Kind:  "Deployment",
			Name:  "sample",
			Image: []string{"docker.io/library/nginx:1.27.0"},
		},
	}

	output := imageClient.setOutput(images)

	assert.Equal(t, [][]string{
		{"Name", "Kind", "Platform", "Image"},
		{"sample", "Deployment", "linux/arm64", "docker.io/library/nginx:1.27.0"},
	}, output)
	assert.Equal(t, "linux/arm64", images[0].Platform)
}
