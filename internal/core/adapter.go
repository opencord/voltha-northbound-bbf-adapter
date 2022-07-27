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
	"context"
	"fmt"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/opencord/voltha-lib-go/v7/pkg/log"
	"github.com/opencord/voltha-northbound-bbf-adapter/internal/clients"
	"github.com/opencord/voltha-protos/v5/go/voltha"
)

var AdapterInstance *VolthaYangAdapter

type VolthaYangAdapter struct {
	volthaNbiClient *clients.VolthaNbiClient
	onosClient      *clients.OnosClient
}

func NewVolthaYangAdapter(nbiClient *clients.VolthaNbiClient, onosClient *clients.OnosClient) *VolthaYangAdapter {
	return &VolthaYangAdapter{
		volthaNbiClient: nbiClient,
		onosClient:      onosClient,
	}
}

func (t *VolthaYangAdapter) GetDevices(ctx context.Context) ([]YangItem, error) {
	devices, err := t.volthaNbiClient.Service.ListDevices(ctx, &empty.Empty{})
	if err != nil {
		return nil, fmt.Errorf("get-devices-failed: %v", err)
	}
	logger.Debugw(ctx, "get-devices-success", log.Fields{"devices": devices})

	items := []YangItem{}

	for _, device := range devices.Items {
		items = append(items, translateDevice(device)...)

		if !device.Root {
			//If the device is an ONU, also expose UNIs
			ports, err := t.volthaNbiClient.Service.ListDevicePorts(ctx, &voltha.ID{Id: device.Id})
			if err != nil {
				return nil, fmt.Errorf("get-onu-ports-failed: %v", err)
			}
			logger.Debugw(ctx, "get-onu-ports-success", log.Fields{"deviceId": device.Id, "ports": ports})

			portsItems, err := translateOnuPorts(device.Id, ports)
			if err != nil {
				logger.Errorw(ctx, "cannot-translate-onu-ports", log.Fields{
					"deviceId": device.Id,
					"err":      err,
				})
				continue
			}

			items = append(items, portsItems...)
		}
	}

	return items, nil
}

func (t *VolthaYangAdapter) GetVlans(ctx context.Context) ([]YangItem, error) {
	services, err := t.onosClient.GetProgrammedSubscribers()
	if err != nil {
		return nil, fmt.Errorf("get-programmed-subscribers-failed: %v", err)
	}
	logger.Debugw(ctx, "get-programmed-subscribers-success", log.Fields{"services": services})

	//No need for other requests if there are no services
	if len(services) == 0 {
		return []YangItem{}, nil
	}

	ports, err := t.onosClient.GetPorts()
	if err != nil {
		return nil, fmt.Errorf("get-onos-ports-failed: %v", err)
	}
	logger.Debugw(ctx, "get-onos-ports-success", log.Fields{"ports": ports})

	items, err := translateVlans(services, ports)
	if err != nil {
		return nil, fmt.Errorf("cannot-translate-vlans: %v", err)
	}

	return items, nil
}

func (t *VolthaYangAdapter) GetBandwidthProfiles(ctx context.Context) ([]YangItem, error) {
	services, err := t.onosClient.GetProgrammedSubscribers()
	if err != nil {
		return nil, fmt.Errorf("get-programmed-subscribers-failed: %v", err)
	}
	logger.Debugw(ctx, "get-programmed-subscribers-success", log.Fields{"services": services})

	//No need for other requests if there are no services
	if len(services) == 0 {
		return []YangItem{}, nil
	}

	bwProfilesMap := map[string]bool{}
	bwProfiles := []clients.BandwidthProfile{}

	for _, service := range services {
		//Get information on downstream bw profile if new
		if _, ok := bwProfilesMap[service.TagInfo.DownstreamBandwidthProfile]; !ok {
			bw, err := t.onosClient.GetBandwidthProfile(service.TagInfo.DownstreamBandwidthProfile)
			if err != nil {
				return nil, fmt.Errorf("get-bw-profile-failed: %s %v", service.TagInfo.DownstreamBandwidthProfile, err)
			}
			logger.Debugw(ctx, "get-bw-profile-success", log.Fields{"bwProfile": bw})

			bwProfiles = append(bwProfiles, *bw)
			bwProfilesMap[service.TagInfo.DownstreamBandwidthProfile] = true
		}

		//Get information on upstream bw profile if new
		if _, ok := bwProfilesMap[service.TagInfo.UpstreamBandwidthProfile]; !ok {
			bw, err := t.onosClient.GetBandwidthProfile(service.TagInfo.UpstreamBandwidthProfile)
			if err != nil {
				return nil, fmt.Errorf("get-bw-profile-failed: %s %v", service.TagInfo.UpstreamBandwidthProfile, err)
			}
			logger.Debugw(ctx, "get-bw-profile-success", log.Fields{"bwProfile": bw})

			bwProfiles = append(bwProfiles, *bw)
			bwProfilesMap[service.TagInfo.UpstreamBandwidthProfile] = true
		}
	}

	items, err := translateBandwidthProfiles(bwProfiles)
	if err != nil {
		return nil, fmt.Errorf("cannot-translate-bandwidth-profiles: %v", err)
	}

	return items, nil
}

func (t *VolthaYangAdapter) GetServices(ctx context.Context) ([]YangItem, error) {
	services, err := t.onosClient.GetProgrammedSubscribers()
	if err != nil {
		return nil, fmt.Errorf("get-programmed-subscribers-failed: %v", err)
	}
	logger.Debugw(ctx, "get-programmed-subscribers-success", log.Fields{"services": services})

	//No need for other requests if there are no services
	if len(services) == 0 {
		return []YangItem{}, nil
	}

	ports, err := t.onosClient.GetPorts()
	if err != nil {
		return nil, fmt.Errorf("get-onos-ports-failed: %v", err)
	}
	logger.Debugw(ctx, "get-onos-ports-success", log.Fields{"ports": ports})

	items, err := translateServices(services, ports)
	if err != nil {
		return nil, fmt.Errorf("cannot-translate-services: %v", err)
	}

	return items, nil
}

func (t *VolthaYangAdapter) ProvisionService(portName string, sTag string, cTag string, technologyProfileId string) error {
	_, err := t.onosClient.ProvisionService(portName, sTag, cTag, technologyProfileId)
	return err
}

func (t *VolthaYangAdapter) RemoveService(portName string, sTag string, cTag string, technologyProfileId string) error {
	_, err := t.onosClient.RemoveService(portName, sTag, cTag, technologyProfileId)
	return err
}
