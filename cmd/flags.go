package cmd

import (
	"github.com/nikhilsbhat/helm-images/pkg"
	"github.com/nikhilsbhat/helm-images/pkg/k8s"
	"github.com/spf13/cobra"
)

// Registers all global flags to utility itself.
func registerFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringArrayVarP(&images.Values, "set", "", nil,
		"set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	cmd.PersistentFlags().StringArrayVarP(&images.StringValues, "set-string", "", nil,
		"set STRING values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	cmd.PersistentFlags().StringArrayVar(&images.FileValues, "set-file", []string{},
		"set values from respective files specified via the command line "+
			"(can specify multiple or separate values with commas: key1=path1,key2=path2)")
	cmd.PersistentFlags().StringArrayVarP(&images.ShowOnly, "show-only", "s", nil,
		"only show manifests rendered from the given templates")
	cmd.PersistentFlags().VarP(&images.ValueFiles, "values", "f",
		"specify values in a YAML file (can specify multiple)")
	cmd.PersistentFlags().StringVarP(&images.Version, "version", "", "",
		"specify a version constraint for the chart version to use, the value passed here would be used to set "+
			"--version for helm template command while generating templates")
	cmd.PersistentFlags().BoolVarP(&images.SkipTests, "skip-tests", "", false,
		"setting this would set '--skip-tests' for helm template command while generating templates")
	cmd.PersistentFlags().BoolVarP(&images.SkipCRDS, "skip-crds", "", false,
		"setting this would set '--skip-crds' for helm template command while generating templates")
	cmd.PersistentFlags().BoolVarP(&images.Validate, "validate", "", false,
		"setting this would set '--validate' for helm template command while generating templates")
}

// Registers all flags to command, get.
func registerGetFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringSliceVarP(&images.Registries, "registry", "r", nil,
		"registry name (docker images belonging to this registry)")
	cmd.PersistentFlags().StringSliceVarP(&images.Skip, "skip", "", nil,
		"list of resources to skip from identifying images, ex: ConfigMap=sample-configmap | configmap=sample-configmap")
	cmd.PersistentFlags().StringSliceVarP(&images.Kind, "kind", "k", k8s.SupportedKinds(),
		"kubernetes app kind to fetch the images from")
	cmd.PersistentFlags().StringVarP(&images.LogLevel, "log-level", "l", "info",
		"log level for the plugin helm images (defaults to info)")
	cmd.PersistentFlags().StringVarP(&images.ImageRegex, "image-regex", "", pkg.ImageRegex,
		"regex used to split helm template rendered")
	cmd.PersistentFlags().BoolVarP(&images.UniqueImages, "unique", "u", false,
		"enable the flag if duplicates to be removed from the retrieved list (disabled by default also overrides --kind)")
	cmd.PersistentFlags().StringVarP(&images.OutputFormat, "output", "o", "",
		"the format to which the output should be rendered to, it should be one of yaml|json|table|csv, if nothing specified it sets to default")
	cmd.PersistentFlags().BoolVarP(&images.FromRelease, "from-release", "", false,
		"enable the flag to fetch the images from release instead (disabled by default)")
	cmd.PersistentFlags().BoolVarP(&images.NoColor, "no-color", "", false,
		"when enabled does not color encode the output")
}
