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
	"time"

	"github.com/opencord/voltha-northbound-bbf-adapter/internal/clients"
	"github.com/opencord/voltha-protos/v5/go/openflow_13"
	"github.com/opencord/voltha-protos/v5/go/voltha"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
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
	assert.Equal(t, "/bbf-device-aggregation:devices/device[name='123145abcdef']", path)
}

func TestDeviceHardwarePath(t *testing.T) {
	path := getDeviceHardwarePath(testDeviceId)
	assert.Equal(t, "/bbf-device-aggregation:devices/device[name='123145abcdef']/data/ietf-hardware:hardware/component[name='123145abcdef']", path)
}

func TestServicePortPath(t *testing.T) {
	path := GetServicePortPath("testService", "testPort")
	assert.Equal(t, "/bbf-nt-service-profile:service-profiles/service-profile[name='testService']/ports/port[name='testPort']", path)
}

func TestVlansPath(t *testing.T) {
	path := GetVlansPath("testProfile")
	assert.Equal(t, "/bbf-l2-access-attributes:vlan-translation-profiles/vlan-translation-profile[name='testProfile']", path)
}

func TestTranslateDevice(t *testing.T) {
	olt := &voltha.Device{
		Id:              testDeviceId,
		Root:            true,
		Vendor:          "BBSim",
		Model:           "asfvolt16",
		SerialNumber:    "BBSIM_OLT_10",
		HardwareVersion: "v0.0.2",
		FirmwareVersion: "v0.0.3",
		AdminState:      voltha.AdminState_ENABLED,
		OperStatus:      voltha.OperStatus_ACTIVE,
	}
	items := translateDevice(olt)

	oltPath := getDevicePath(testDeviceId)
	oltHwPath := getDeviceHardwarePath(testDeviceId)

	expected := []YangItem{
		{
			Path:  oltPath + "/type",
			Value: DeviceTypeOlt,
		},
		{
			Path:  oltHwPath + "/mfg-name",
			Value: "BBSim",
		},
		{
			Path:  oltHwPath + "/model-name",
			Value: "asfvolt16",
		},
		{
			Path:  oltHwPath + "/hardware-rev",
			Value: "v0.0.2",
		},
		{
			Path:  oltHwPath + "/firmware-rev",
			Value: "v0.0.3",
		},
		{
			Path:  oltHwPath + "/serial-num",
			Value: "BBSIM_OLT_10",
		},
		{
			Path:  oltHwPath + "/state/admin-state",
			Value: ietfAdminStateUnlocked,
		},
		{
			Path:  oltHwPath + "/state/oper-state",
			Value: ietfOperStateEnabled,
		},
	}

	assert.NotEmpty(t, items, "No OLT items")
	for _, e := range expected {
		val, ok := getItemWithPath(items, e.Path)
		assert.True(t, ok, e.Path+" missing for OLT")
		assert.Equal(t, e.Value, val, "Wrong value for "+e.Path)
	}

	onu := &voltha.Device{
		Id:              testDeviceId,
		Root:            false,
		Vendor:          "BBSM",
		Model:           "v0.0.1",
		SerialNumber:    "BBSM000a0001",
		HardwareVersion: "v0.0.2",
		FirmwareVersion: "v0.0.3",
		AdminState:      voltha.AdminState_ENABLED,
		OperStatus:      voltha.OperStatus_ACTIVE,
		ParentId:        "abcdef1234",
		ParentPortNo:    1,
	}
	items = translateDevice(onu)

	onuPath := getDevicePath(testDeviceId)
	onuHwPath := getDeviceHardwarePath(testDeviceId)

	expected = []YangItem{
		{
			Path:  onuPath + "/type",
			Value: DeviceTypeOnu,
		},
		{
			Path:  onuHwPath + "/mfg-name",
			Value: "BBSM",
		},
		{
			Path:  onuHwPath + "/model-name",
			Value: "v0.0.1",
		},
		{
			Path:  onuHwPath + "/hardware-rev",
			Value: "v0.0.2",
		},
		{
			Path:  onuHwPath + "/firmware-rev",
			Value: "v0.0.3",
		},
		{
			Path:  onuHwPath + "/serial-num",
			Value: "BBSM000a0001",
		},
		{
			Path:  onuHwPath + "/state/admin-state",
			Value: ietfAdminStateUnlocked,
		},
		{
			Path:  onuHwPath + "/state/oper-state",
			Value: ietfOperStateEnabled,
		},
		{
			Path:  onuHwPath + "/parent",
			Value: "abcdef1234",
		},
		{
			Path:  onuHwPath + "/parent-rel-pos",
			Value: "1",
		},
	}

	assert.NotEmpty(t, items, "No ONU items")
	for _, e := range expected {
		val, ok := getItemWithPath(items, e.Path)
		assert.True(t, ok, e.Path+" missing for ONU")
		assert.Equal(t, e.Value, val, "Wrong value for "+e.Path)
	}
}

func TestTranslateOnuPorts(t *testing.T) {
	ports := &voltha.Ports{
		Items: []*voltha.Port{
			{
				PortNo:     0,
				Type:       voltha.Port_ETHERNET_UNI,
				OperStatus: voltha.OperStatus_ACTIVE,
			},
		},
	}

	_, err := translateOnuPorts(testDeviceId, ports)
	assert.Error(t, err, "No error for missing Ofp port")

	ports = &voltha.Ports{
		Items: []*voltha.Port{
			{
				PortNo: 0,
				Type:   voltha.Port_ETHERNET_UNI,
				OfpPort: &openflow_13.OfpPort{
					Name: "BBSM000a0001-1",
				},
				OperStatus: voltha.OperStatus_ACTIVE,
			},
			{
				PortNo: 1,
				Type:   voltha.Port_ETHERNET_UNI,
				OfpPort: &openflow_13.OfpPort{
					Name: "BBSM000a0001-2",
				},
				OperStatus: voltha.OperStatus_UNKNOWN,
			},
			{
				PortNo:     0,
				Type:       voltha.Port_PON_ONU,
				OperStatus: voltha.OperStatus_UNKNOWN,
			},
		},
	}

	portsItems, err := translateOnuPorts(testDeviceId, ports)
	assert.Nil(t, err, "Translation error")

	/*2 items for 2 UNIs, PON is ignored*/
	assert.Equal(t, 4, len(portsItems), "No ports items")

	interfacesPath := getDevicePath(testDeviceId) + "/data/ietf-interfaces:interfaces"

	expected := []YangItem{
		{
			Path:  fmt.Sprintf("%s/interface[name='%s']/oper-status", interfacesPath, "BBSM000a0001-1"),
			Value: ietfOperStateUp,
		},
		{
			Path:  fmt.Sprintf("%s/interface[name='%s']/type", interfacesPath, "BBSM000a0001-1"),
			Value: "bbf-xpon-if-type:onu-v-vrefpoint",
		},
		{
			Path:  fmt.Sprintf("%s/interface[name='%s']/oper-status", interfacesPath, "BBSM000a0001-2"),
			Value: ietfOperStateUnknown,
		},
		{
			Path:  fmt.Sprintf("%s/interface[name='%s']/type", interfacesPath, "BBSM000a0001-2"),
			Value: "bbf-xpon-if-type:onu-v-vrefpoint",
		},
	}

	for _, e := range expected {
		val, ok := getItemWithPath(portsItems, e.Path)
		assert.True(t, ok, e.Path+" missing for ports")
		assert.Equal(t, e.Value, val, "Wrong value for "+e.Path)
	}
}

func TestTranslateOnuActive(t *testing.T) {
	timestamp := time.Now()
	eventHeader := &voltha.EventHeader{
		Id:       "Voltha.openolt.ONU_ACTIVATED.1657705515351182767",
		RaisedTs: timestamppb.New(timestamp),
		Category: voltha.EventCategory_EQUIPMENT,
		Type:     voltha.EventType_DEVICE_EVENT,
	}

	deviceEvent := &voltha.DeviceEvent{
		ResourceId:      testDeviceId,
		DeviceEventName: "ONU_ACTIVATED_RAISE_EVENT",
		Description:     "ONU Event - ONU_ACTIVATED - Raised",
		Context:         map[string]string{},
	}

	_, _, err := TranslateOnuActivatedEvent(eventHeader, deviceEvent)
	assert.Error(t, err, "Empty context produces no error")

	deviceEvent.Context[eventContextKeyPonId] = "0"
	deviceEvent.Context[eventContextKeyOnuSn] = "BBSM000a0001"
	deviceEvent.Context[eventContextKeyOltSn] = "BBSIM_OLT_10"

	notificationPath := "/bbf-xpon-onu-states:onu-state-change"
	expected := []YangItem{
		{
			Path:  notificationPath + "/detected-serial-number",
			Value: "BBSM000a0001",
		},
		{
			Path:  notificationPath + "/onu-state-last-change",
			Value: timestamp.Format(time.RFC3339),
		},
		{
			Path:  notificationPath + "/onu-state",
			Value: "bbf-xpon-onu-types:onu-present",
		},
		{
			Path:  notificationPath + "/detected-registration-id",
			Value: testDeviceId,
		},
	}

	notificationItems, channelTerminationItems, err := TranslateOnuActivatedEvent(eventHeader, deviceEvent)
	assert.Nil(t, err, "Translation error")

	assert.NotEmpty(t, channelTerminationItems, "No channel termination items")

	assert.NotEmpty(t, notificationItems, "No notification items")

	_, ok := getItemWithPath(notificationItems, notificationPath+"/channel-termination-ref")
	assert.True(t, ok, "No channel termination reference in notification")

	for _, e := range expected {
		val, ok := getItemWithPath(notificationItems, e.Path)
		assert.True(t, ok, e.Path+" missing for notification")
		assert.Equal(t, e.Value, val, "Wrong value for "+e.Path)
	}
}

func TestTranslateServices(t *testing.T) {
	subscriber := clients.ProgrammedSubscriber{
		Location: "of:00001/256",
		TagInfo: clients.SadisUniTag{
			UniTagMatch:                 100,
			PonCTag:                     4096,
			PonSTag:                     102,
			TechnologyProfileID:         64,
			UpstreamBandwidthProfile:    "BW1",
			DownstreamBandwidthProfile:  "BW2",
			UpstreamOltBandwidthProfile: "OLTBW",
			IsDhcpRequired:              true,
			IsIgmpRequired:              false,
			IsPPPoERequired:             false,
			ConfiguredMacAddress:        "00:11:22:33:44:55",
			EnableMacLearning:           true,
			UsPonCTagPriority:           1,
			UsPonSTagPriority:           2,
			DsPonCTagPriority:           3,
			DsPonSTagPriority:           -1,
			ServiceName:                 "testService",
		},
	}

	alias := ServiceAlias{
		Key: ServiceKey{
			Port: "TESTPORT-1",
			STag: "101",
			CTag: "102",
			TpId: "64",
		},
		ServiceName: "TESTPORT-1-testService",
		VlansName:   "TESTPORT-1-testService-vlans",
	}

	servicesItesm, err := translateService(subscriber.TagInfo, alias)
	assert.Nil(t, err, "Translation error")

	assert.NotEmpty(t, servicesItesm, "No services items")

	servicePortPath := ServiceProfilesPath + "/service-profile[name='TESTPORT-1-testService']/ports/port[name='TESTPORT-1']"

	expected := []YangItem{
		{
			Path:  servicePortPath + "/bbf-nt-service-profile-voltha:configured-mac-address",
			Value: "00:11:22:33:44:55",
		},
		{
			Path:  servicePortPath + "/bbf-nt-service-profile-voltha:mac-learning-enabled",
			Value: "true",
		},
		{
			Path:  servicePortPath + "/bbf-nt-service-profile-voltha:dhcp-required",
			Value: "true",
		},
		{
			Path:  servicePortPath + "/bbf-nt-service-profile-voltha:igmp-required",
			Value: "false",
		},
		{
			Path:  servicePortPath + "/bbf-nt-service-profile-voltha:pppoe-required",
			Value: "false",
		},
	}

	_, ok := getItemWithPath(servicesItesm, servicePortPath+"/port-vlans/port-vlan[name='TESTPORT-1-testService-vlans']")
	assert.True(t, ok, "No vlans leafref in services")

	_, ok = getItemWithPath(servicesItesm, servicePortPath+"/bbf-nt-service-profile-voltha:downstream-olt-bp-name")
	assert.False(t, ok, "Downstream OLT bandwidth profile should not be present")

	for _, e := range expected {
		val, ok := getItemWithPath(servicesItesm, e.Path)
		assert.True(t, ok, e.Path+" missing for services")
		assert.Equal(t, e.Value, val, "Wrong value for "+e.Path)
	}
}

func TestTranslateVlans(t *testing.T) {
	subscriber := clients.ProgrammedSubscriber{
		Location: "of:00001/256",
		TagInfo: clients.SadisUniTag{
			UniTagMatch:                 100,
			PonCTag:                     4096,
			PonSTag:                     102,
			TechnologyProfileID:         64,
			UpstreamBandwidthProfile:    "BW1",
			DownstreamBandwidthProfile:  "BW2",
			UpstreamOltBandwidthProfile: "OLTBW",
			IsDhcpRequired:              true,
			IsIgmpRequired:              false,
			IsPPPoERequired:             false,
			ConfiguredMacAddress:        "00:11:22:33:44:55",
			EnableMacLearning:           true,
			UsPonCTagPriority:           1,
			UsPonSTagPriority:           2,
			DsPonCTagPriority:           3,
			DsPonSTagPriority:           -1,
			ServiceName:                 "testService",
		},
	}

	alias := ServiceAlias{
		Key: ServiceKey{
			Port: "TESTPORT-1",
			STag: "101",
			CTag: "102",
			TpId: "64",
		},
		ServiceName: "TESTPORT-1-testService",
		VlansName:   "TESTPORT-1-testService-vlans",
	}

	vlanItems, err := translateVlans(subscriber.TagInfo, alias)
	assert.Nil(t, err, "Translation error")

	assert.NotEmpty(t, vlanItems, "No vlans items")

	vlanPath := VlansPath + "/vlan-translation-profile[name='TESTPORT-1-testService-vlans']"

	expected := []YangItem{
		{
			Path:  vlanPath + "/match-criteria/outer-tag/vlan-id",
			Value: "100",
		},
		{
			Path:  vlanPath + "/ingress-rewrite/push-second-tag/vlan-id",
			Value: "any",
		},
		{
			Path:  vlanPath + "/ingress-rewrite/push-outer-tag/vlan-id",
			Value: "102",
		},
		{
			Path:  vlanPath + "/match-criteria/second-tag/vlan-id",
			Value: "any",
		},
		{
			Path:  vlanPath + "/ingress-rewrite/push-second-tag/pbit",
			Value: "1",
		},
		{
			Path:  vlanPath + "/ingress-rewrite/push-outer-tag/pbit",
			Value: "2",
		},
		{
			Path:  vlanPath + "/ingress-rewrite/push-second-tag/bbf-voltha-vlan-translation:dpbit",
			Value: "3",
		},
	}

	_, ok := getItemWithPath(vlanItems, vlanPath+"/ingress-rewrite/push-outer-tag/bbf-voltha-vlan-translation:dpbit")
	assert.False(t, ok, "Pbit value should not be present")

	for _, e := range expected {
		val, ok := getItemWithPath(vlanItems, e.Path)
		assert.True(t, ok, e.Path+" missing for vlans")
		assert.Equal(t, e.Value, val, "Wrong value for "+e.Path)
	}
}
