package cmd

import (
	"github.com/nikhilsbhat/helm-images/pkg"
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
	cmd.PersistentFlags().StringSliceVarP(&images.Kind, "kind", "k", nil,
		"kubernetes app kind to fetch the images from (if not specified all kinds are considered)")
	cmd.PersistentFlags().StringVarP(&images.ImageRegex, "image-regex", "", pkg.ImageRegex, "regex used to split helm template rendered")
}
