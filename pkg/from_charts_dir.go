package pkg

import (
	"fmt"
	"os"
	"path/filepath"

	imageErr "github.com/nikhilsbhat/helm-images/pkg/errors"
)

type chartInfo struct {
	name string
	path string
}

// getChartsFromDir discovers all helm charts in the specified directory.
func (image *Images) getChartsFromDir() ([]chartInfo, error) {
	chartsDir := image.ChartsDir
	image.log.Debugf("scanning directory '%s' for helm charts", chartsDir)

	// Check if directory exists
	info, err := os.Stat(chartsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &imageErr.ImageError{Message: fmt.Sprintf("directory '%s' does not exist", chartsDir)}
		}

		return nil, fmt.Errorf("failed to access directory '%s': %w", chartsDir, err)
	}

	if !info.IsDir() {
		return nil, &imageErr.ImageError{Message: fmt.Sprintf("'%s' is not a directory", chartsDir)}
	}

	charts := make([]chartInfo, 0)

	// Read directory entries
	entries, err := os.ReadDir(chartsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory '%s': %w", chartsDir, err)
	}

	// Look for Chart.yaml in each subdirectory
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		chartPath := filepath.Join(chartsDir, entry.Name())
		chartYamlPath := filepath.Join(chartPath, "Chart.yaml")

		// Check if Chart.yaml exists
		if _, err := os.Stat(chartYamlPath); err == nil {
			image.log.Debugf("discovered helm chart: %s at %s", entry.Name(), chartPath)
			charts = append(charts, chartInfo{
				name: entry.Name(),
				path: chartPath,
			})
		}
	}

	if len(charts) == 0 {
		image.log.Warnf("no helm charts found in directory '%s'", chartsDir)
	} else {
		image.log.Infof("found %d helm chart(s) in directory '%s'", len(charts), chartsDir)
	}

	return charts, nil
}

// getChartManifestFromDir renders a single chart from the charts directory.
func (image *Images) getChartManifestFromDir(chartPath, chartName string) ([]byte, error) {
	image.log.Debugf("rendering helm chart from path '%s'", chartPath)

	// Temporarily set the chart path and release name
	originalChart := image.chart
	originalRelease := image.release
	image.chart = chartPath
	image.release = chartName
	defer func() {
		image.chart = originalChart
		image.release = originalRelease
	}()

	// Use existing template rendering logic
	return image.getChartFromTemplate()
}
