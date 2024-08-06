package pkg

import (
	"github.com/thoas/go-funk"
	"helm.sh/helm/v3/pkg/release"
)

type releasesToSkip []skipReleaseInfo

func (skips releasesToSkip) filterRelease(releases []*release.Release) []*release.Release {
	return funk.Filter(releases, func(release *release.Release) bool {
		for _, skip := range skips {
			if skip.name == release.Name && skip.namespace == release.Namespace {
				return false
			}
		}

		return true
	}).([]*release.Release)
}
