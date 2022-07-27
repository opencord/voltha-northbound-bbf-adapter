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
	"strconv"
	"time"

	"github.com/opencord/voltha-northbound-bbf-adapter/internal/clients"
	"github.com/opencord/voltha-protos/v5/go/common"
	"github.com/opencord/voltha-protos/v5/go/voltha"
)

const (
	DeviceAggregationModule = "bbf-device-aggregation"
	DevicesPath             = "/" + DeviceAggregationModule + ":devices"

	ServiceProfileModule = "bbf-nt-service-profile"
	ServiceProfilesPath  = "/" + ServiceProfileModule + ":service-profiles"

	VlansModule = "bbf-l2-access-attributes"
	VlansPath   = "/" + VlansModule + ":vlan-translation-profiles"

	BandwidthProfileModule = "bbf-nt-line-profile"
	BandwidthProfilesPath  = "/" + BandwidthProfileModule + ":line-bandwidth-profiles"

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

	//Values to allow any VLAN ID
	YangVlanIdAny   = "any"
	VolthaVlanIdAny = 4096
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

//GetServicePortPath returns the yang path to a service's port node
func GetServicePortPath(serviceName string, portName string) string {
	return fmt.Sprintf("%s/service-profile[name='%s']/ports/port[name='%s']", ServiceProfilesPath, serviceName, portName)
}

//GetVlansPath returns the yang path to a vlan translation profile's root node
func GetVlansPath(serviceName string) string {
	return fmt.Sprintf("%s/vlan-translation-profile[name='%s']", VlansPath, serviceName)
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
		result = append(result, []YangItem{
			{
				Path:  devicePath + "/type",
				Value: DeviceTypeOnu,
			},
			{
				Path:  hardwarePath + "/parent",
				Value: device.ParentId,
			},
			{
				Path:  hardwarePath + "/parent-rel-pos",
				Value: strconv.FormatUint(uint64(device.ParentPortNo), 10),
			},
		}...)
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

//translateServices returns a slice of yang items that represent the currently programmed services
func translateServices(subscribers []clients.ProgrammedSubscriber, ports []clients.OnosPort) ([]YangItem, error) {
	//Create a map of port IDs to port names
	//e.g. of:00000a0a0a0a0a0a/256 to BBSM000a0001-1
	portNames := map[string]string{}

	for _, port := range ports {
		portId := fmt.Sprintf("%s/%s", port.Element, port.Port)
		name, ok := port.Annotations["portName"]
		if ok {
			portNames[portId] = name
		}
	}

	result := []YangItem{}

	for _, subscriber := range subscribers {
		portName, ok := portNames[subscriber.Location]
		if !ok {
			return nil, fmt.Errorf("no-port-name-for-location: %s", subscriber.Location)
		}

		serviceName := fmt.Sprintf("%s-%s", portName, subscriber.TagInfo.ServiceName)

		portPath := GetServicePortPath(serviceName, portName)

		if subscriber.TagInfo.ConfiguredMacAddress != "" {
			result = append(result, YangItem{
				Path:  portPath + "/bbf-nt-service-profile-voltha:configured-mac-address",
				Value: subscriber.TagInfo.ConfiguredMacAddress,
			})
		}

		result = append(result, []YangItem{
			{
				Path:  fmt.Sprintf("%s/port-vlans/port-vlan[name='%s']", portPath, serviceName),
				Value: "",
			},
			{
				Path:  portPath + "/bbf-nt-service-profile-voltha:technology-profile-id",
				Value: strconv.Itoa(subscriber.TagInfo.TechnologyProfileID),
			},
			{
				Path:  portPath + "/bbf-nt-service-profile-voltha:downstream-subscriber-bp-name",
				Value: subscriber.TagInfo.DownstreamBandwidthProfile,
			},
			{
				Path:  portPath + "/bbf-nt-service-profile-voltha:upstream-subscriber-bp-name",
				Value: subscriber.TagInfo.UpstreamBandwidthProfile,
			},
			{
				Path:  portPath + "/bbf-nt-service-profile-voltha:mac-learning-enabled",
				Value: strconv.FormatBool(subscriber.TagInfo.EnableMacLearning),
			},
			{
				Path:  portPath + "/bbf-nt-service-profile-voltha:dhcp-required",
				Value: strconv.FormatBool(subscriber.TagInfo.IsDhcpRequired),
			},
			{
				Path:  portPath + "/bbf-nt-service-profile-voltha:igmp-required",
				Value: strconv.FormatBool(subscriber.TagInfo.IsIgmpRequired),
			},
			{
				Path:  portPath + "/bbf-nt-service-profile-voltha:pppoe-required",
				Value: strconv.FormatBool(subscriber.TagInfo.IsPPPoERequired),
			},
		}...)

		if subscriber.TagInfo.UpstreamOltBandwidthProfile != "" {
			result = append(result, YangItem{
				Path:  portPath + "/bbf-nt-service-profile-voltha:upstream-olt-bp-name",
				Value: subscriber.TagInfo.UpstreamOltBandwidthProfile,
			})
		}

		if subscriber.TagInfo.DownstreamOltBandwidthProfile != "" {
			result = append(result, YangItem{
				Path:  portPath + "/bbf-nt-service-profile-voltha:downstream-olt-bp-name",
				Value: subscriber.TagInfo.UpstreamOltBandwidthProfile,
			})
		}
	}

	return result, nil
}

//translateVlans returns a slice of yang items that represent the vlans used by programmed services
func translateVlans(subscribers []clients.ProgrammedSubscriber, ports []clients.OnosPort) ([]YangItem, error) {
	//Create a map of port IDs to port names
	//e.g. of:00000a0a0a0a0a0a/256 to BBSM000a0001-1
	portNames := map[string]string{}

	for _, port := range ports {
		portId := fmt.Sprintf("%s/%s", port.Element, port.Port)
		name, ok := port.Annotations["portName"]
		if ok {
			portNames[portId] = name
		}
	}

	result := []YangItem{}

	for _, subscriber := range subscribers {
		portName, ok := portNames[subscriber.Location]
		if !ok {
			return nil, fmt.Errorf("no-port-name-for-location: %s", subscriber.Location)
		}

		serviceName := fmt.Sprintf("%s-%s", portName, subscriber.TagInfo.ServiceName)

		vlansPath := GetVlansPath(serviceName)

		uniTagMatch := YangVlanIdAny
		sTag := YangVlanIdAny
		cTag := YangVlanIdAny

		if subscriber.TagInfo.UniTagMatch != VolthaVlanIdAny {
			uniTagMatch = strconv.Itoa(subscriber.TagInfo.UniTagMatch)
		}
		if subscriber.TagInfo.PonSTag != VolthaVlanIdAny {
			sTag = strconv.Itoa(subscriber.TagInfo.PonSTag)
		}
		if subscriber.TagInfo.PonCTag != VolthaVlanIdAny {
			cTag = strconv.Itoa(subscriber.TagInfo.PonCTag)
		}

		if subscriber.TagInfo.UniTagMatch > 0 {
			result = append(result, []YangItem{
				{
					Path:  vlansPath + "/match-criteria/outer-tag/vlan-id",
					Value: uniTagMatch,
				},
				{
					Path:  vlansPath + "/match-criteria/second-tag/vlan-id",
					Value: "any",
				},
			}...)
		}

		if subscriber.TagInfo.UsPonSTagPriority >= 0 {
			result = append(result, YangItem{
				Path:  vlansPath + "/ingress-rewrite/push-outer-tag/pbit",
				Value: strconv.Itoa(subscriber.TagInfo.UsPonSTagPriority),
			})
		}
		if subscriber.TagInfo.DsPonSTagPriority >= 0 {
			result = append(result, YangItem{
				Path:  vlansPath + "/ingress-rewrite/push-outer-tag/bbf-voltha-vlan-translation:dpbit",
				Value: strconv.Itoa(subscriber.TagInfo.DsPonSTagPriority),
			})
		}
		if subscriber.TagInfo.UsPonCTagPriority >= 0 {
			result = append(result, YangItem{
				Path:  vlansPath + "/ingress-rewrite/push-second-tag/pbit",
				Value: strconv.Itoa(subscriber.TagInfo.UsPonCTagPriority),
			})
		}
		if subscriber.TagInfo.DsPonCTagPriority >= 0 {
			result = append(result, YangItem{
				Path:  vlansPath + "/ingress-rewrite/push-second-tag/bbf-voltha-vlan-translation:dpbit",
				Value: strconv.Itoa(subscriber.TagInfo.DsPonCTagPriority),
			})
		}

		result = append(result, []YangItem{
			{
				Path:  vlansPath + "/ingress-rewrite/push-outer-tag/vlan-id",
				Value: sTag,
			},
			{
				Path:  vlansPath + "/ingress-rewrite/push-second-tag/vlan-id",
				Value: cTag,
			},
		}...)
	}

	return result, nil
}

//translateBandwidthProfiles returns a slice of yang items that represent the bandwidth profiles used by programmed services
func translateBandwidthProfiles(bwProfiles []clients.BandwidthProfile) ([]YangItem, error) {
	result := []YangItem{}

	//TODO: The best way to translate this information is still under discussion, but the code
	// to retrieve it is ready. Since this is not fundamental at the moment, an empty slice is
	// returned, and the correct translation can be added here at a later time.

	return result, nil
}
