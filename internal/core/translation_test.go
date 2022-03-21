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
	"fmt"
	"testing"

	"github.com/opencord/voltha-protos/v5/go/voltha"
	"github.com/stretchr/testify/assert"
)

const (
	testDeviceId = "123145abcdef"
)

func getItemWithPath(items []YangItem, path string) (value string, ok bool) {
	for _, item := range items {
		if item.Path == path {
			return item.Value, true
		}
	}

	return "", false
}

func TestDevicePath(t *testing.T) {
	path := getDevicePath(testDeviceId)
	assert.Equal(t, fmt.Sprintf("/bbf-device-aggregation:devices/device[name='%s']", testDeviceId), path)
}

func TestTranslateDevice(t *testing.T) {
	olt := voltha.Device{
		Id:   testDeviceId,
		Root: true,
	}
	items := translateDevice(olt)

	val, ok := getItemWithPath(items, fmt.Sprintf("%s/type", getDevicePath(testDeviceId)))
	assert.True(t, ok, "No type item for olt")
	assert.Equal(t, DeviceTypeOlt, val)

	onu := voltha.Device{
		Id:   testDeviceId,
		Root: false,
	}
	items = translateDevice(onu)

	val, ok = getItemWithPath(items, fmt.Sprintf("%s/type", getDevicePath(testDeviceId)))
	assert.True(t, ok, "No type item for onu")
	assert.Equal(t, DeviceTypeOnu, val)
}

func TestTranslateDevices(t *testing.T) {
	devicesNum := 10

	//Create test devices
	devices := voltha.Devices{
		Items: []*voltha.Device{},
	}

	for i := 0; i < devicesNum; i++ {
		devices.Items = append(devices.Items, &voltha.Device{
			Id:   fmt.Sprintf("%d", i),
			Root: i%2 == 0,
		})
	}

	//Translate them to items
	items := translateDevices(devices)

	//Check if the number of generated items is correct
	singleDeviceItemsNum := len(translateDevice(*devices.Items[0]))
	assert.Equal(t, singleDeviceItemsNum*devicesNum, len(items))

	//Check if the content is right
	for i := 0; i < devicesNum; i++ {
		val, ok := getItemWithPath(items, fmt.Sprintf("%s/type", getDevicePath(devices.Items[i].Id)))
		assert.True(t, ok, fmt.Sprintf("No type item for device %d", i))

		if devices.Items[i].Root {
			assert.Equal(t, DeviceTypeOlt, val)
		} else {
			assert.Equal(t, DeviceTypeOnu, val)
		}
	}
}
