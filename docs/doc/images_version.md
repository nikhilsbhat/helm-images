## images version

Command to fetch the version of helm-images installed

### Synopsis

This will help user to find what version of helm-images plugin he/she installed in her machine.

```
images version [flags]
```

### Options

```
  -h, --help   help for version
```

### Options inherited from parent commands

```
      --revision int             revision of your release from which the images to be fetched
      --set stringArray          set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)
      --set-file stringArray     set values from respective files specified via the command line (can specify multiple or separate values with commas: key1=path1,key2=path2)
      --set-string stringArray   set STRING values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)
  -s, --show-only stringArray    only show manifests rendered from the given templates
      --skip-crds                setting this would set '--skip-crds' for helm template command while generating templates
      --skip-tests               setting this would set '--skip-tests' for helm template command while generating templates
      --validate                 setting this would set '--validate' for helm template command while generating templates
  -f, --values ValueFiles        specify values in a YAML file (can specify multiple) (default [])
      --version string           specify a version constraint for the chart version to use, the value passed here would be used to set --version for helm template command while generating templates
```

### SEE ALSO

* [images](images.md)	 - Utility that helps in fetching images which are part of deployment

###### Auto generated by spf13/cobra on 12-Oct-2024
