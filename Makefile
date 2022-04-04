# Copyright 2022-present Open Networking Foundation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# set default shell
SHELL = bash -e -o pipefail

# Variables
VERSION                  ?= $(shell cat ./VERSION)

DOCKER_LABEL_VCS_DIRTY     = false
ifneq ($(shell git ls-files --others --modified --exclude-standard 2>/dev/null | wc -l | sed -e 's/ //g'),0)
    DOCKER_LABEL_VCS_DIRTY = true
endif
## Docker related
DOCKER_EXTRA_ARGS        ?=
DOCKER_REGISTRY          ?=
DOCKER_REPOSITORY        ?=
DOCKER_TAG               ?= ${VERSION}$(shell [[ ${DOCKER_LABEL_VCS_DIRTY} == "true" ]] && echo "-dirty" || true)
ADAPTER_IMAGENAME        := ${DOCKER_REGISTRY}${DOCKER_REPOSITORY}voltha-northbound-bbf-adapter:${DOCKER_TAG}
DOCKER_TARGET            ?= prod
TYPE                     ?= minimal

## Docker labels. Only set ref and commit date if committed
DOCKER_LABEL_VCS_URL       ?= $(shell git remote get-url $(shell git remote))
DOCKER_LABEL_VCS_REF       = $(shell git rev-parse HEAD)
DOCKER_LABEL_BUILD_DATE    ?= $(shell date -u "+%Y-%m-%dT%H:%M:%SZ")
DOCKER_LABEL_COMMIT_DATE   = $(shell git show -s --format=%cd --date=iso-strict HEAD)

DOCKER_BUILD_ARGS ?= \
	${DOCKER_EXTRA_ARGS} \
	--build-arg org_label_schema_version="${VERSION}" \
	--build-arg org_label_schema_vcs_url="${DOCKER_LABEL_VCS_URL}" \
	--build-arg org_label_schema_vcs_ref="${DOCKER_LABEL_VCS_REF}" \
	--build-arg org_label_schema_build_date="${DOCKER_LABEL_BUILD_DATE}" \
	--build-arg org_opencord_vcs_commit_date="${DOCKER_LABEL_COMMIT_DATE}" \
	--build-arg org_opencord_vcs_dirty="${DOCKER_LABEL_VCS_DIRTY}"

# tool containers
VOLTHA_TOOLS_VERSION ?= 2.5.3

# Dependencies versions
LIBYANG_VERSION		?= f9bbd46fa3a6b09291ec0c1e93ebf569f3e30bd1
SYSREPO_VERSION		?= 64e3c66442d682c31e979db289c4c64a3ec1f6c1
LIBNETCONF2_VERSION	?= ef7d3e3ca1504e8ca9c4f4b5dd3847ba17bb809d
NETOPEER2_VERSION	?= 39800066f9fbbde9b55e6cfde77927eeb5627c83

# Default user and password for the netconf user in docker-build
NETCONF_USER ?= voltha
NETCONF_PASSWORD ?= onf

# This container is built to include the necessary sysrepo libraries
# to succesfully build and test the code in this repository
BUILDER_IMAGE_AND_TAG ?= voltha/bbf-adapter-builder:local 
build-tools: build/tools/Dockerfile.builder
	docker build \
	  -t ${BUILDER_IMAGE_AND_TAG} \
	  -f build/tools/Dockerfile.builder . \
	  --build-arg LIBYANG_VERSION=${LIBYANG_VERSION} \
	  --build-arg SYSREPO_VERSION=${SYSREPO_VERSION}

GO_LOCAL_BUILDER            = docker run --rm --user $$(id -u):$$(id -g) -v ${CURDIR}:/app $(shell test -t 0 && echo "-it") -v gocache:/.cache -v gocache-${VOLTHA_TOOLS_VERSION}:/go/pkg ${BUILDER_IMAGE_AND_TAG} go
GO                          = docker run --rm --user $$(id -u):$$(id -g) -v ${CURDIR}:/app $(shell test -t 0 && echo "-it") -v gocache:/.cache -v gocache-${VOLTHA_TOOLS_VERSION}:/go/pkg voltha/voltha-ci-tools:${VOLTHA_TOOLS_VERSION}-golang go
GO_JUNIT_REPORT             = docker run --rm --user $$(id -u):$$(id -g) -v ${CURDIR}:/app -i voltha/voltha-ci-tools:${VOLTHA_TOOLS_VERSION}-go-junit-report go-junit-report
GOCOVER_COBERTURA           = docker run --rm --user $$(id -u):$$(id -g) -v ${CURDIR}:/app/src/github.com/opencord/voltha-northbound-bbf-adapter -v gocache:/.cache -v gocache-${VOLTHA_TOOLS_VERSION}:/go/pkg voltha/voltha-ci-tools:${VOLTHA_TOOLS_VERSION}-gocover-cobertura gocover-cobertura
GOLANGCI_LINT_LOCAL_BUILDER = docker run --rm --user $$(id -u):$$(id -g) -v ${CURDIR}:/app $(shell test -t 0 && echo "-it") -v gocache:/.cache -v gocache-${VOLTHA_TOOLS_VERSION}:/go/pkg ${BUILDER_IMAGE_AND_TAG} golangci-lint
HADOLINT                    = docker run --rm --user $$(id -u):$$(id -g) -v ${CURDIR}:/app $(shell test -t 0 && echo "-it") voltha/voltha-ci-tools:${VOLTHA_TOOLS_VERSION}-hadolint hadolint

.PHONY: docker-build local-protos local-lib-go help test sca
.DEFAULT_GOAL := help

help: ## Print help for each Makefile target
	@echo
	@echo Northbound BBF Adapter
	@echo
	@echo Translates the BBF yang model to VOLTHA Northbound APIs
	@echo
	@echo "Usage: make [<target>]"
	@echo "where available targets are:"
	@grep '^[[:alpha:]_-]*:.* ##' $(MAKEFILE_LIST) \
		| sort | awk 'BEGIN {FS=":.* ## "}; {printf "%-25s : %s\n", $$1, $$2};'

## Local Development Helpers
local-protos: ## Copies a local version of the voltha-protos dependency into the vendor directory
ifdef LOCAL_PROTOS
	rm -rf vendor/github.com/opencord/voltha-protos/v5/go
	mkdir -p vendor/github.com/opencord/voltha-protos/v5/go
	cp -r ${LOCAL_PROTOS}/go/* vendor/github.com/opencord/voltha-protos/v5/go
	rm -rf vendor/github.com/opencord/voltha-protos/v5/go/vendor
endif

local-lib-go: ## Copies a local version of the voltha-lib-go dependency into the vendor directory
ifdef LOCAL_LIB_GO
	mkdir -p vendor/github.com/opencord/voltha-lib-go/v7/pkg
	cp -r ${LOCAL_LIB_GO}/pkg/* vendor/github.com/opencord/voltha-lib-go/v7/pkg/
endif

## Docker targets
build: docker-build ## Alias for 'docker build'

docker-build: local-lib-go build-tools ## Build the BBF Adapter docker container
	docker build \
	  -t ${DOCKER_REGISTRY}${DOCKER_REPOSITORY}voltha-northbound-bbf-adapter:${DOCKER_TAG} \
	  -f build/package/Dockerfile.bbf-adapter . \
	  --build-arg NETCONF_USER=${NETCONF_USER} \
	  --build-arg NETCONF_PASSWORD=${NETCONF_PASSWORD} \
	  --build-arg LIBNETCONF2_VERSION=${LIBNETCONF2_VERSION} \
	  --build-arg NETOPEER2_VERSION=${NETOPEER2_VERSION}

docker-push: ## Push the docker images to an external repository
	docker push ${ADAPTER_IMAGENAME}
ifdef BUILD_PROFILED
	docker push ${ADAPTER_IMAGENAME}-profile
endif
ifdef BUILD_RACE
	docker push ${ADAPTER_IMAGENAME}-rd
endif

docker-kind-load: ## Load docker images into a KinD cluster
	@if [ "`kind get clusters | grep voltha-$(TYPE)`" = '' ]; then echo "no voltha-$(TYPE) cluster found" && exit 1; fi
	kind load docker-image ${ADAPTER_IMAGENAME} --name=voltha-$(TYPE) --nodes $(shell kubectl get nodes --template='{{range .items}}{{.metadata.name}},{{end}}' | sed 's/,$$//')

test: build-tools ## Run unit tests
	@mkdir -p ./tests/results
	@${GO_LOCAL_BUILDER} test -mod=vendor -v -coverprofile ./tests/results/go-test-coverage.out -covermode count ./... 2>&1 | tee ./tests/results/go-test-results.out ;\
	RETURN=$$? ;\
	${GO_JUNIT_REPORT} < ./tests/results/go-test-results.out > ./tests/results/go-test-results.xml ;\
	${GOCOVER_COBERTURA} < ./tests/results/go-test-coverage.out > ./tests/results/go-test-coverage.xml ;\
	exit $$RETURN

lint: local-lib-go lint-mod lint-dockerfile ## Run all lint targets

lint-dockerfile: ## Perform static analysis on Dockerfile
	@echo "Running Dockerfile lint check..."
	@${HADOLINT} $$(find ./build -name "Dockerfile*")
	@echo "Dockerfile lint check OK"

lint-mod: ## Verify the Go dependencies
	@echo "Running dependency check..."
	@${GO} mod verify
	@echo "Dependency check OK. Running vendor check..."
	@git status > /dev/null
	@git diff-index --quiet HEAD -- go.mod go.sum vendor || (echo "ERROR: Staged or modified files must be committed before running this test" && git status -- go.mod go.sum vendor && exit 1)
	@[[ `git ls-files --exclude-standard --others go.mod go.sum vendor` == "" ]] || (echo "ERROR: Untracked files must be cleaned up before running this test" && git status -- go.mod go.sum vendor && exit 1)
	${GO} mod tidy
	${GO} mod vendor
	@git status > /dev/null
	@git diff-index --quiet HEAD -- go.mod go.sum vendor || (echo "ERROR: Modified files detected after running go mod tidy / go mod vendor" && git status -- go.mod go.sum vendor && git checkout -- go.mod go.sum vendor && exit 1)
	@[[ `git ls-files --exclude-standard --others go.mod go.sum vendor` == "" ]] || (echo "ERROR: Untracked files detected after running go mod tidy / go mod vendor" && git status -- go.mod go.sum vendor && git checkout -- go.mod go.sum vendor && exit 1)
	@echo "Vendor check OK."

sca: build-tools ## Runs static code analysis with the golangci-lint tool
	@rm -rf ./sca-report
	@mkdir -p ./sca-report
	@echo "Running static code analysis..."
	@${GOLANGCI_LINT_LOCAL_BUILDER} run --deadline=6m --out-format junit-xml ./... | tee ./sca-report/sca-report.xml
	@echo ""
	@echo "Static code analysis OK"

clean: distclean ## Removes any local filesystem artifacts generated by a build

distclean: ## Removes any local filesystem artifacts generated by a build or test run
	rm -rf ./sca-report

mod-update: ## Update go mod files
	${GO} mod tidy
	${GO} mod vendor