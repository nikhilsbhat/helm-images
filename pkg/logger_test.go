package pkg_test

import (
	"testing"

	"github.com/nikhilsbhat/helm-images/pkg"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestGetLoglevel(t *testing.T) {
	t.Run("should return warn level", func(t *testing.T) {
		actual := pkg.GetLoglevel("warning")
		assert.Equal(t, log.WarnLevel, actual)
	})
	t.Run("should return trace level", func(t *testing.T) {
		actual := pkg.GetLoglevel("trace")
		assert.Equal(t, log.TraceLevel, actual)
	})
	t.Run("should return debug level", func(t *testing.T) {
		actual := pkg.GetLoglevel("debug")
		assert.Equal(t, log.DebugLevel, actual)
	})
	t.Run("should return fatal level", func(t *testing.T) {
		actual := pkg.GetLoglevel("fatal")
		assert.Equal(t, log.FatalLevel, actual)
	})
	t.Run("should return error level", func(t *testing.T) {
		actual := pkg.GetLoglevel("error")
		assert.Equal(t, log.ErrorLevel, actual)
	})
}
