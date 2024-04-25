# Helm Images


[![Go Report Card](https://goreportcard.com/badge/github.com/nikhilsbhat/helm-images)](https://goreportcard.com/report/github.com/nikhilsbhat/helm-images) 
[![shields](https://img.shields.io/badge/license-MIT-blue)](https://github.com/nikhilsbhat/helm-images/blob/master/LICENSE) 
[![shields](https://godoc.org/github.com/nikhilsbhat/helm-images?status.svg)](https://godoc.org/github.com/nikhilsbhat/helm-images)
[![shields](https://img.shields.io/github/v/tag/nikhilsbhat/helm-images.svg)](https://github.com/nikhilsbhat/helm-images/tags)
[![shields](https://img.shields.io/github/downloads/nikhilsbhat/helm-images/total.svg)](https://github.com/nikhilsbhat/helm-images/releases)
[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/images)](https://artifacthub.io/packages/search?repo=images)


This helm plugins helps in identifying all images that would be part of helm chart deployment.

## Introduction

Identifying all images just before the deployment of the helm chart is not a straight-forward task.

This Helm plugin was created to ease this task. This can be installed as an add-on to the helm.

It helps in filtering images based on the Kubernetes type. It also helps in filtering images based on a registry that it is part of.

```shell
helm images get prometheus-standalone ~/prometheus-setup/prometheus-standalone -f ~/prometheus-setup/prometheus-standalone/values-standalone-1.yaml
# executing above command would yield results something like below:
quay.io/prometheus/node-exporter:v1.1.2
k8s.gcr.io/kube-state-metrics/kube-state-metrics:v2.0.0
quay.io/prometheus/alertmanager:v0.21.0
prom/pushgateway:v1.3.1
jimmidyson/configmap-reload:v0.5.0
quay.io/prometheus/alertmanager:v0.21.0

# using the same plugin can list images which are part of specified release
helm images get prometheus-standalone --from-release --registry quay.io
# above command should fetch all the images from a helm release 'prometheus-standalone' by limiting to registry 'quay.io', which results as below:
quay.io/prometheus/alertmanager:v0.21.0
quay.io/prometheus/alertmanager:v0.21.0
```
## Installation

```shell
helm plugin install https://github.com/nikhilsbhat/helm-images
```
Use the executable just like any other go-cli application.

## Usage

```bash
helm images [command] [flags]
```
Make sure appropriate command is used for the actions, to check the available commands and flags use `helm images --help`

```bash
Lists all images that would be part of helm deployment.

Usage:
  images [command] [flags]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  get         Fetches all images those are part of specified chart/release
  help        Help about any command
  version     Command to fetch the version of helm-images installed

Flags:
  -h, --help                     help for images
      --set stringArray          set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)
      --set-file stringArray     set values from respective files specified via the command line (can specify multiple or separate values with commas: key1=path1,key2=path2)
      --set-string stringArray   set STRING values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)
  -s, --show-only stringArray    only show manifests rendered from the given templates
      --skip-crds                setting this would set '--skip-crds' for helm template command while generating templates
      --skip-tests               setting this would set '--skip-tests' for helm template command while generating templates
      --validate                 setting this would set '--validate' for helm template command while generating templates
  -f, --values ValueFiles        specify values in a YAML file (can specify multiple) (default [])
      --version string           specify a version constraint for the chart version to use, the value passed here would be used to set --version for helm template command while generating templates


Use "images [command] --help" for more information about a command.
```

## Commands
### `get`

```shell
Lists all images those are part of specified chart/release and matches the pattern or part of specified registry.

Usage:
  images get [RELEASE] [CHART] [flags]

Examples:
  helm images get prometheus-standalone path/to/chart/prometheus-standalone -f ~/path/to/override-config.yaml
  helm images get prometheus-standalone --from-release --registry quay.io -o table
  helm images get prometheus-standalone --from-release --registry quay.io --unique
  helm images get prometheus-standalone --from-release --registry quay.io -o yaml
  helm images get oci://registry-1.docker.io/bitnamicharts/airflow -o yaml
  helm images get kong-2.35.0.tgz -o json

Flags:
      --from-release         enable the flag to fetch the images from release instead (disabled by default)
  -h, --help                 help for get
      --image-regex string   regex used to split helm template rendered (default "---\\n# Source:\\s.*.")
  -k, --kind strings         kubernetes app kind to fetch the images from (default [Deployment,StatefulSet,DaemonSet,CronJob,Job,ReplicaSet,Pod,Alertmanager,Prometheus,ThanosRuler,Grafana,Thanos,Receiver,ConfigMap])
  -l, --log-level string     log level for the plugin helm images (defaults to info) (default "info")
      --no-color             when enabled does not color encode the output
  -o, --output string        the format to which the output should be rendered to, it should be one of yaml|json|table|csv, if nothing specified it sets to default
  -r, --registry strings     registry name (docker images belonging to this registry)
      --skip strings         list of resources to skip from identifying images, ex: ConfigMap=sample-configmap | configmap=sample-configmap
  -u, --unique               enable the flag if duplicates to be removed from the retrieved list (disabled by default also overrides --kind)

Global Flags:
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

## Documentation

Updated documentation on all available commands and flags can be found [here](https://github.com/nikhilsbhat/helm-images/blob/master/docs/doc/images.md).

## Caveats

If the plugin is not listing the expected images, then most likely the `helm images plugin` does not support fetching images from the `workload` that it is part of.</br>
Invoking the plugin with log-level set to `debug` should give information if the plugin is not supporting the workload.

The plugin only supports the resources that are defined under flag [--kind](https://github.com/nikhilsbhat/helm-images/blob/master/cmd/flags.go#L37).

Available resources can be found [here](https://github.com/nikhilsbhat/helm-images/blob/master/pkg/k8s/k8s.go#L23).
