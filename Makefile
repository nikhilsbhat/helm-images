GOFMT_FILES?=$(shell find . -not -path "./vendor/*" -type f -name '*.go')
APP_NAME?=helm-images
APP_DIR?=$(shell git rev-parse --show-toplevel)
DEV?=${DEVBOX_TRUE}
SRC_PACKAGES=$(shell go list -mod=vendor ./... | grep -v "vendor" | grep -v "mocks")
BUILD_ENVIRONMENT?=${ENVIRONMENT}
VERSION?=0.1.0
REVISION?=$(shell git rev-parse --verify HEAD)
DATE?=$(shell date)
PlATFORM?=$(shell go env GOOS)
ARCHITECTURE?=$(shell go env GOARCH)
GOVERSION?=$(shell go version | awk '{printf $$3}')
BUILD_WITH_FLAGS="-s -w -X 'github.com/nikhilsbhat/helm-images/version.Version=${VERSION}' -X 'github.com/nikhilsbhat/helm-images/version.Env=${BUILD_ENVIRONMENT}' -X 'github.com/nikhilsbhat/helm-images/version.BuildDate=${DATE}' -X 'github.com/nikhilsbhat/helm-images/version.Revision=${REVISION}' -X 'github.com/nikhilsbhat/helm-images/version.Platform=${PlATFORM}/${ARCHITECTURE}' -X 'github.com/nikhilsbhat/helm-images/version.GoVersion=${GOVERSION}'"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

.PHONY: help
help: ## Prints help (only for targets with comments)
	@grep -E '^[a-zA-Z0-9._-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

local.fmt: ## Lints all the go code in the application.
	gofmt -w $(GOFMT_FILES)
	$(GOBIN)/goimports -w $(GOFMT_FILES)

local.check: local.fmt ## Loads all the dependencies to vendor directory
	go mod vendor
	go mod tidy

local.build: local.check ## Generates the artifact with the help of 'go build'
	GOVERSION=${GOVERSION} BUILD_ENVIRONMENT=${BUILD_ENVIRONMENT} goreleaser build --rm-dist

local.push: local.build ## Pushes built artifact to the specified location

local.run: local.build ## Generates the artifact and start the service in the current directory
	./${APP_NAME}

publish: local.check ## Builds and publishes the app
	GOVERSION=${GOVERSION} BUILD_ENVIRONMENT=${BUILD_ENVIRONMENT} goreleaser release --rm-dist

lint: ## Lint's application for errors, it is a linters aggregator (https://github.com/golangci/golangci-lint).
	if [ -z "${DEV}" ]; then golangci-lint run --color always ; else docker run --rm -v $(APP_DIR):/app -w /app golangci/golangci-lint:v1.31-alpine golangci-lint run --color always ; fi

report: ## Publishes the go-report of the appliction (uses go-reportcard)
	if [ -z "${DEV}" ]; then goreportcard-cli -v ; else docker run --rm -v $(APP_DIR):/app -w /app basnik/goreportcard-cli:latest goreportcard-cli -v ; fi

dev.prerequisite.up: ## Sets up the development environment with all necessary components.
	$(APP_DIR)/scripts/prerequisite.sh

dev.prerequisite.purge: ## Teardown the development environment by removing all components.

install.hooks: ## install pre-push hooks for the repository.
	${APP_DIR}/scripts/hook.sh ${APP_DIR}

generate.mock: ## generates mocks for the selected source packages.
	@go generate ${SRC_PACKAGES}