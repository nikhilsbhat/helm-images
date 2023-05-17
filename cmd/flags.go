package cmd

import (
	"github.com/nikhilsbhat/helm-images/pkg"
	"github.com/nikhilsbhat/helm-images/pkg/k8s"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/spf13/cobra"
)

// Registers all global flags to utility itself.
func registerFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringArrayVar(&images.Values, "set", []string{},
		"set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	cmd.PersistentFlags().StringArrayVar(&images.StringValues, "set-string", []string{},
		"set STRING values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	cmd.PersistentFlags().StringArrayVar(&images.FileValues, "set-file", []string{},
		"set values from respective files specified via the command line (can specify multiple or separate values with commas: key1=path1,key2=path2)") //nolint:lll
	cmd.PersistentFlags().VarP(&images.ValueFiles, "values", "f",
		"specify values in a YAML file (can specify multiple)")
}

// Registers all flags to command, get.
func registerGetFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringSliceVarP(&images.Registries, "registry", "r", nil,
		"registry name (docker images belonging to this registry)")
	cmd.PersistentFlags().StringSliceVarP(&images.Kind, "kind", "k", []string{
		k8s.KindDeployment, k8s.KindStatefulSet, k8s.KindDaemonSet,
		k8s.KindCronJob, k8s.KindJob, k8s.KindReplicaSet, k8s.KindPod,
		monitoringv1.AlertmanagersKind, monitoringv1.PrometheusesKind, monitoringv1.ThanosRulerKind,
		k8s.KindGrafana,
	},
		"kubernetes app kind to fetch the images from")
	cmd.PersistentFlags().StringVarP(&images.LogLevel, "log-level", "l", "info",
		"log level for the plugin helm images (defaults to info)")
	cmd.PersistentFlags().StringVarP(&images.ImageRegex, "image-regex", "", pkg.ImageRegex,
		"regex used to split helm template rendered")
	cmd.PersistentFlags().BoolVarP(&images.UniqueImages, "unique", "u", false,
		"enable the flag if duplicates to be removed from the retrieved list (disabled by default also overrides --kind)")
	cmd.PersistentFlags().BoolVarP(&images.JSON, "json", "j", false,
		"enable the flag to display images retrieved in json format (disabled by default)")
	cmd.PersistentFlags().BoolVarP(&images.YAML, "yaml", "y", false,
		"enable the flag to display images retrieved in yaml format (disabled by default)")
	cmd.PersistentFlags().BoolVarP(&images.Table, "table", "t", false,
		"enable the flag to display images retrieved in table format (disabled by default)")
	cmd.PersistentFlags().BoolVarP(&images.FromRelease, "from-release", "", false,
		"enable the flag to fetch the images from release instead (disabled by default)")
}
