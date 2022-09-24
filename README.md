# Helm Images


[![Go Report Card](https://goreportcard.com/badge/github.com/nikhilsbhat/helm-images)](https://goreportcard.com/report/github.com/nikhilsbhat/helm-images) 
[![shields](https://img.shields.io/badge/license-MIT-blue)](https://github.com/nikhilsbhat/helm-images/blob/master/LICENSE) 
[![shields](https://godoc.org/github.com/nikhilsbhat/helm-images?status.svg)](https://godoc.org/github.com/nikhilsbhat/helm-images)


This helm plugins helps in identifying all images that would be part of helm chart deployment.

## Introduction

Identifying all images just before the deployment of the helm chart is not a straightforward task.

To make it simple, the helm plugin is leveraged. This can be installed as an add-on to the helm.

It helps in filtering images based on the Kubernetes type. It also helps in filtering images based on a registry which it is part of.

```shell
helm images get prometheus-standalone ~/prometheus-setup/prometheus-standalone -f ~/prometheus-setup/prometheus-standalone/values-standalone-1.yaml
# executing above command would yield results something like below:
quay.io/prometheus/node-exporter:v1.1.2
k8s.gcr.io/kube-state-metrics/kube-state-metrics:v2.0.0
quay.io/prometheus/alertmanager:v0.21.0
prom/pushgateway:v1.3.1
jimmidyson/configmap-reload:v0.5.0
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
Lists all images that would be part of helm deployment would be listed.

Usage:
  images [command] [flags]

Available Commands:
  completion  generate the autocompletion script for the specified shell
  get         Fetches all images part of deployment
  help        Help about any command
  version     Command to fetch the version of helm-images installed

Flags:
  -h, --help                     help for images
      --set stringArray          set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)
      --set-file stringArray     set values from respective files specified via the command line (can specify multiple or separate values with commas: key1=path1,key2=path2)
      --set-string stringArray   set STRING values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)
  -f, --values ValueFiles        specify values in a YAML file (can specify multiple) (default [])


Use "images [command] --help" for more information about a command.
```

## Commands
### `get`

```shell
Lists all images that matches the pattern or part of specified registry.

Usage:
  images get [RELEASE] [CHART] [flags]

Flags:
  -h, --help                 help for get
      --image-regex string   regex used to split helm template rendered (default "---\\n# Source:\\s.*.")
  -j, --json                 enable the flag display information retrieved in json format (disabled by default)
  -k, --kind strings         kubernetes app kind to fetch the images from (default [Deployment,StatefulSet,DaemonSet,CronJob,Job,ReplicaSet])
  -r, --registry strings     registry name (docker images belonging to this registry)
  -u, --unique               enable the flag if duplicates to be removed from the images that are retrieved (disabled by default also overrides --kind)
  -y, --yaml                 enable the flag display information retrieved in yaml format (disabled by default)

Global Flags:
      --set stringArray          set values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)
      --set-file stringArray     set values from respective files specified via the command line (can specify multiple or separate values with commas: key1=path1,key2=path2)
      --set-string stringArray   set STRING values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)
  -f, --values ValueFiles        specify values in a YAML file (can specify multiple) (default [])
```