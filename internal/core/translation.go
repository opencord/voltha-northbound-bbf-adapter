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

	"github.com/opencord/voltha-protos/v5/go/voltha"
)

const (
	DeviceAggregationModel = "bbf-device-aggregation"
	DevicesPath            = "/" + DeviceAggregationModel + ":devices"
	DeviceTypeOlt          = "bbf-device-types:olt"
	DeviceTypeOnu          = "bbf-device-types:onu"
)

type YangItem struct {
	Path  string
	Value string
}

//getDevicePath returns the yang path to the root of the device with a specific ID
func getDevicePath(id string) string {
	return fmt.Sprintf("%s/device[name='%s']", DevicesPath, id)
}

//translateDevice returns a slice of yang items that represent a voltha device
func translateDevice(device voltha.Device) []YangItem {
	devicePath := getDevicePath(device.Id)

	typeItem := YangItem{}
	typeItem.Path = fmt.Sprintf("%s/type", devicePath)

	if device.Root {
		typeItem.Value = DeviceTypeOlt
	} else {
		typeItem.Value = DeviceTypeOnu
	}

	return []YangItem{typeItem}
}

//translateDevices returns a slice of yang items that represent a list of voltha devices
func translateDevices(devices voltha.Devices) []YangItem {
	result := []YangItem{}

	for _, device := range devices.Items {
		result = append(result, translateDevice(*device)...)
	}

	return result
}
