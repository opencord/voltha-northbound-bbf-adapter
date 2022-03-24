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

FROM --platform=linux/amd64 golang:1.16.3-alpine3.13 AS dev

RUN mkdir -m 777 /.cache /go/pkg

RUN apk add --no-cache build-base=0.5-r2 pcre2-dev=10.36-r0 git=2.30.2-r0 cmake=3.18.4-r1

#Install golangci-lint
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.44.2


ARG LIBYANG_VERSION
ARG SYSREPO_VERSION

#Build libyang
WORKDIR /
RUN git clone https://github.com/CESNET/libyang.git
WORKDIR /libyang
RUN git checkout $LIBYANG_VERSION && mkdir build
WORKDIR /libyang/build
RUN cmake -D CMAKE_BUILD_TYPE:String="Release" .. && \
    make && \
    make install

#Build sysrepo
WORKDIR /
RUN git clone https://github.com/sysrepo/sysrepo.git
WORKDIR /sysrepo
RUN git checkout $SYSREPO_VERSION && mkdir build
WORKDIR /sysrepo/build
RUN cmake -D CMAKE_BUILD_TYPE:String="Release" .. && \
    make && \
    make install

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