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

	"github.com/opencord/voltha-protos/v5/go/common"
	"github.com/opencord/voltha-protos/v5/go/voltha"
)

const (
	DeviceAggregationModel = "bbf-device-aggregation"
	DevicesPath            = "/" + DeviceAggregationModel + ":devices"
	HardwarePath           = "data/ietf-hardware:hardware"

	//Device types
	DeviceTypeOlt = "bbf-device-types:olt"
	DeviceTypeOnu = "bbf-device-types:onu"

	//Admin states
	ietfAdminStateUnknown  = "unknown"
	ietfAdminStateLocked   = "locked"
	ietfAdminStateUnlocked = "unlocked"

	//Oper states
	ietfOperStateUnknown  = "unknown"
	ietfOperStateDisabled = "disabled"
	ietfOperStateEnabled  = "enabled"
	ietfOperStateTesting  = "testing"
)

type YangItem struct {
	Path  string
	Value string
}

//getDevicePath returns the yang path to the root of the device with a specific ID
func getDevicePath(id string) string {
	return fmt.Sprintf("%s/device[name='%s']", DevicesPath, id)
}

//getDevicePath returns the yang path to the root of the device's hardware module in its data mountpoint
func getDeviceHardwarePath(id string) string {
	return fmt.Sprintf("%s/device[name='%s']/%s/component[name='%s']", DevicesPath, id, HardwarePath, id)
}

//ietfHardwareAdminState returns the string that represents the ietf-hardware admin state
//enum value corresponding to the one of VOLTHA
func ietfHardwareAdminState(volthaAdminState voltha.AdminState_Types) string {
	//TODO: verify this mapping is correct
	switch volthaAdminState {
	case common.AdminState_UNKNOWN:
		return ietfAdminStateUnknown
	case common.AdminState_PREPROVISIONED:
	case common.AdminState_DOWNLOADING_IMAGE:
	case common.AdminState_ENABLED:
		return ietfAdminStateUnlocked
	case common.AdminState_DISABLED:
		return ietfAdminStateLocked
	}

	//TODO: does something map to "shutting-down" ?

	return ietfAdminStateUnknown
}

//ietfHardwareOperState returns the string that represents the ietf-hardware oper state
//enum value corresponding to the one of VOLTHA
func ietfHardwareOperState(volthaOperState voltha.OperStatus_Types) string {
	//TODO: verify this mapping is correct
	switch volthaOperState {
	case common.OperStatus_UNKNOWN:
		return ietfOperStateUnknown
	case common.OperStatus_TESTING:
		return ietfOperStateTesting
	case common.OperStatus_ACTIVE:
		return ietfOperStateEnabled
	case common.OperStatus_DISCOVERED:
	case common.OperStatus_ACTIVATING:
	case common.OperStatus_FAILED:
	case common.OperStatus_RECONCILING:
	case common.OperStatus_RECONCILING_FAILED:
		return ietfOperStateDisabled
	}

	return ietfOperStateUnknown
}

//translateDevice returns a slice of yang items that represent a voltha device
func translateDevice(device voltha.Device) []YangItem {
	devicePath := getDevicePath(device.Id)
	hardwarePath := getDeviceHardwarePath(device.Id)

	result := []YangItem{}

	//Device type
	if device.Root {
		result = append(result, YangItem{
			Path:  fmt.Sprintf("%s/type", devicePath),
			Value: DeviceTypeOlt,
		})
	} else {
		result = append(result, YangItem{
			Path:  fmt.Sprintf("%s/type", devicePath),
			Value: DeviceTypeOnu,
		})
	}

	//Vendor name
	result = append(result, YangItem{
		Path:  fmt.Sprintf("%s/mfg-name", hardwarePath),
		Value: device.Vendor,
	})

	//Model
	result = append(result, YangItem{
		Path:  fmt.Sprintf("%s/model-name", hardwarePath),
		Value: device.Model,
	})

	//Hardware version
	result = append(result, YangItem{
		Path:  fmt.Sprintf("%s/hardware-rev", hardwarePath),
		Value: device.HardwareVersion,
	})

	//Firmware version
	result = append(result, YangItem{
		Path:  fmt.Sprintf("%s/firmware-rev", hardwarePath),
		Value: device.FirmwareVersion,
	})

	//Serial number
	result = append(result, YangItem{
		Path:  fmt.Sprintf("%s/serial-num", hardwarePath),
		Value: device.SerialNumber,
	})

	//Administrative state
	//Translates VOLTHA admin state enum to ietf-hardware enum
	result = append(result, YangItem{
		Path:  fmt.Sprintf("%s/state/admin-state", hardwarePath),
		Value: ietfHardwareAdminState(device.AdminState),
	})

	//Operative state
	result = append(result, YangItem{
		Path:  fmt.Sprintf("%s/state/oper-state", hardwarePath),
		Value: ietfHardwareOperState(device.OperStatus),
	})

	return result
}

//translateDevices returns a slice of yang items that represent a list of voltha devices
func translateDevices(devices voltha.Devices) []YangItem {
	result := []YangItem{}

	for _, device := range devices.Items {
		result = append(result, translateDevice(*device)...)
	}

	return result
}
