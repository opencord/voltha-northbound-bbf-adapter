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

# Dockerfile with golang and the sysrepo dependencies for voltha-northbound-bff-adapter
# This image is used for testing, static code analysis and building

FROM --platform=linux/amd64 golang:1.16.3-buster AS dev

RUN mkdir -m 777 /.cache /go/pkg

#Install golangci-lint
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.44.2

RUN \
    apt-get update && apt-get install -y libpcre2-dev=10.32-5 --no-install-recommends && \
    rm -rf /var/lib/apt/lists/*

# Download and install library packages
RUN \
    wget https://github.com/CESNET/libyang/releases/download/v2.0.112/libyang2_2.0.112.1-1_amd64.deb -O libyang2.deb && \
    wget https://github.com/CESNET/libyang/releases/download/v2.0.112/libyang2-dev_2.0.112.1-1_amd64.deb -O libyang2-dev.deb && \
    wget https://github.com/sysrepo/sysrepo/releases/download/v2.0.53/libsysrepo6_2.0.53.1-1_amd64.deb -O libsysrepo6.deb && \
    wget https://github.com/sysrepo/sysrepo/releases/download/v2.0.53/libsysrepo-dev_2.0.53.1-1_amd64.deb -O libsysrepo-dev.deb

RUN dpkg -i libyang2.deb libyang2-dev.deb libsysrepo6.deb libsysrepo-dev.deb

RUN rm libyang2.deb libyang2-dev.deb libsysrepo6.deb libsysrepo-dev.deb

WORKDIR /app

ARG org_label_schema_version=unknown
ARG org_label_schema_vcs_url=unknown
ARG org_label_schema_vcs_ref=unknown
ARG org_label_schema_build_date=unknown
ARG org_opencord_vcs_commit_date=unknown
ARG org_opencord_vcs_dirty=unknown

LABEL \
org.label-schema.schema-version=1.0 \
org.label-schema.name=voltha-northbound-bbf-adapter \
org.label-schema.version=$org_label_schema_version \
org.label-schema.vcs-url=$org_label_schema_vcs_url \
org.label-schema.vcs-ref=$org_label_schema_vcs_ref \
org.label-schema.build-date=$org_label_schema_build_date \
org.opencord.vcs-commit-date=$org_opencord_vcs_commit_date \
org.opencord.vcs-dirty=$org_opencord_vcs_dirty