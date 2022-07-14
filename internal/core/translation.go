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
	"time"

	"github.com/opencord/voltha-protos/v5/go/common"
	"github.com/opencord/voltha-protos/v5/go/voltha"
)

const (
	DeviceAggregationModule = "bbf-device-aggregation"
	DevicesPath             = "/" + DeviceAggregationModule + ":devices"

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
	ietfOperStateUp       = "up"
	ietfOperStateDown     = "down"

	//Keys of useful values in device events
	eventContextKeyPonId = "pon-id"
	eventContextKeyOnuSn = "serial-number"
	eventContextKeyOltSn = "olt-serial-number"
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
	return fmt.Sprintf("%s/device[name='%s']/data/ietf-hardware:hardware/component[name='%s']", DevicesPath, id, id)
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

//ietfHardwareOperState returns the string that represents the ietf-interfaces oper state
//enum value corresponding to the one of VOLTHA
func ietfInterfacesOperState(volthaOperState voltha.OperStatus_Types) string {
	//TODO: verify this mapping is correct
	switch volthaOperState {
	case common.OperStatus_UNKNOWN:
		return ietfOperStateUnknown
	case common.OperStatus_TESTING:
		return ietfOperStateTesting
	case common.OperStatus_ACTIVE:
		return ietfOperStateUp
	case common.OperStatus_DISCOVERED:
	case common.OperStatus_ACTIVATING:
	case common.OperStatus_FAILED:
	case common.OperStatus_RECONCILING:
	case common.OperStatus_RECONCILING_FAILED:
		return ietfOperStateDown
	}

	return ietfOperStateUnknown
}

//translateDevice returns a slice of yang items that represent a voltha device
func translateDevice(device *voltha.Device) []YangItem {
	devicePath := getDevicePath(device.Id)
	hardwarePath := getDeviceHardwarePath(device.Id)

	result := []YangItem{}

	//Device type
	if device.Root {
		//OLT
		result = append(result, YangItem{
			Path:  devicePath + "/type",
			Value: DeviceTypeOlt,
		})
	} else {
		//ONU
		result = append(result, YangItem{
			Path:  devicePath + "/type",
			Value: DeviceTypeOnu,
		})
	}

	//Vendor name
	result = append(result, YangItem{
		Path:  hardwarePath + "/mfg-name",
		Value: device.Vendor,
	})

	//Model
	result = append(result, YangItem{
		Path:  hardwarePath + "/model-name",
		Value: device.Model,
	})

	//Hardware version
	result = append(result, YangItem{
		Path:  hardwarePath + "/hardware-rev",
		Value: device.HardwareVersion,
	})

	//Firmware version
	result = append(result, YangItem{
		Path:  hardwarePath + "/firmware-rev",
		Value: device.FirmwareVersion,
	})

	//Serial number
	result = append(result, YangItem{
		Path:  hardwarePath + "/serial-num",
		Value: device.SerialNumber,
	})

	//Administrative state
	//Translates VOLTHA admin state enum to ietf-hardware enum
	result = append(result, YangItem{
		Path:  hardwarePath + "/state/admin-state",
		Value: ietfHardwareAdminState(device.AdminState),
	})

	//Operative state
	result = append(result, YangItem{
		Path:  hardwarePath + "/state/oper-state",
		Value: ietfHardwareOperState(device.OperStatus),
	})

	return result
}

//translateOnuPorts returns a slice of yang items that represent the UNIs of an ONU
func translateOnuPorts(deviceId string, ports *voltha.Ports) ([]YangItem, error) {
	interfacesPath := getDevicePath(deviceId) + "/data/ietf-interfaces:interfaces"
	result := []YangItem{}

	for _, port := range ports.Items {
		if port.Type == voltha.Port_ETHERNET_UNI {
			if port.OfpPort == nil {
				return nil, fmt.Errorf("no-ofp-port-in-uni: %s %d", deviceId, port.PortNo)
			}

			interfacePath := fmt.Sprintf("%s/interface[name='%s']", interfacesPath, port.OfpPort.Name)

			result = append(result, []YangItem{
				{
					Path:  interfacePath + "/type",
					Value: "bbf-xpon-if-type:onu-v-vrefpoint",
				},
				{
					Path:  interfacePath + "/oper-status",
					Value: ietfInterfacesOperState(port.OperStatus),
				},
			}...)
		}
	}

	return result, nil
}

//TranslateOnuActivatedEvent returns a slice of yang items and the name of the channel termination to populate
//an ONU discovery notification with data from ONU_ACTIVATED_RAISE_EVENT coming from the Kafka bus
func TranslateOnuActivatedEvent(eventHeader *voltha.EventHeader, deviceEvent *voltha.DeviceEvent) (notification []YangItem, channelTermination []YangItem, err error) {

	//TODO: the use of this notification, which requires the creation of a dummy channel termination node,
	//is temporary, and will be substituted with a more fitting one as soon as it will be defined

	//Check if the needed information is present
	ponId, ok := deviceEvent.Context[eventContextKeyPonId]
	if !ok {
		return nil, nil, fmt.Errorf("missing-key-from-event-context: %s", eventContextKeyPonId)
	}
	oltId, ok := deviceEvent.Context[eventContextKeyOltSn]
	if !ok {
		return nil, nil, fmt.Errorf("missing-key-from-event-context: %s", eventContextKeyPonId)
	}
	ponName := oltId + "-pon-" + ponId

	onuSn, ok := deviceEvent.Context[eventContextKeyOnuSn]
	if !ok {
		return nil, nil, fmt.Errorf("missing-key-from-event-context: %s", eventContextKeyOnuSn)
	}

	notificationPath := "/bbf-xpon-onu-states:onu-state-change"

	notification = []YangItem{
		{
			Path:  notificationPath + "/detected-serial-number",
			Value: onuSn,
		},
		{
			Path:  notificationPath + "/channel-termination-ref",
			Value: ponName,
		},
		{
			Path:  notificationPath + "/onu-state-last-change",
			Value: eventHeader.RaisedTs.AsTime().Format(time.RFC3339),
		},
		{
			Path:  notificationPath + "/onu-state",
			Value: "bbf-xpon-onu-types:onu-present",
		},
		{
			Path:  notificationPath + "/detected-registration-id",
			Value: deviceEvent.ResourceId,
		},
	}

	channelTermination = []YangItem{
		{
			Path:  fmt.Sprintf("/ietf-interfaces:interfaces/interface[name='%s']/type", ponName),
			Value: "bbf-if-type:vlan-sub-interface",
		},
	}

	return notification, channelTermination, nil
}
