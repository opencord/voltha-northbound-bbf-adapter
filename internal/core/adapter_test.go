/*
* Copyright 2022-present Open Networking Foundation

* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at

* http://www.apache.org/licenses/LICENSE-2.0

* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package core

import (
	"testing"

	"github.com/opencord/voltha-northbound-bbf-adapter/internal/clients"
	"github.com/stretchr/testify/assert"
)

func TestLocationsToPortsMap(t *testing.T) {
	ports := []clients.OnosPort{
		{
			Element: "of:00001",
			Port:    "256",
			Annotations: map[string]string{
				"portName": "TESTPORT-1",
			},
		},
		{
			Element: "of:00001",
			Port:    "257",
			Annotations: map[string]string{
				"portName": "TESTPORT-2",
			},
		},
	}

	portNames := getLocationsToPortsMap(ports)

	assert.NotEmpty(t, portNames, "Empty map")

	name, ok := portNames["of:00001/256"]
	assert.True(t, ok, "First port name not found")
	assert.Equal(t, "TESTPORT-1", name)

	name, ok = portNames["of:00001/257"]
	assert.True(t, ok, "Second port name not found")
	assert.Equal(t, "TESTPORT-2", name)
}

func TestServiceAliasKVPath(t *testing.T) {
	serviceKey := ServiceKey{
		Port: "PORT",
		STag: "100",
		CTag: "101",
		TpId: "64",
	}

	path := getServiceAliasKVPath(serviceKey)

	assert.Equal(t, "services/PORT/100/101/64", path)
}
