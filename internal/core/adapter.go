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
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/opencord/voltha-lib-go/v7/pkg/db"
	"github.com/opencord/voltha-lib-go/v7/pkg/db/kvstore"
	"github.com/opencord/voltha-lib-go/v7/pkg/log"
	"github.com/opencord/voltha-northbound-bbf-adapter/internal/clients"
	"github.com/opencord/voltha-protos/v5/go/voltha"
)

var AdapterInstance *VolthaYangAdapter

const (
	kvStoreServices = "services"
)

type VolthaYangAdapter struct {
	volthaNbiClient *clients.VolthaNbiClient
	onosClient      *clients.OnosClient
	kvStore         *db.Backend
}

func NewVolthaYangAdapter(nbiClient *clients.VolthaNbiClient, onosClient *clients.OnosClient, kvBackend *db.Backend) *VolthaYangAdapter {
	return &VolthaYangAdapter{
		volthaNbiClient: nbiClient,
		onosClient:      onosClient,
		kvStore:         kvBackend,
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

func getLocationsToPortsMap(ports []clients.OnosPort) map[string]string {
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

	return portNames
}

func (t *VolthaYangAdapter) getServiceAliasOrFallback(ctx context.Context, uniTagServiceName string, key ServiceKey) (*ServiceAlias, error) {
	alias, err := t.LoadServiceAlias(ctx, key)
	if err != nil {
		//Happens in case a service is provisioned using ONOS directly,
		//bypassing the adapter
		serviceName := fmt.Sprintf("%s-%s", key.Port, uniTagServiceName)
		alias = &ServiceAlias{
			Key:         key,
			ServiceName: serviceName,
			VlansName:   serviceName + "-vlans",
		}

		logger.Warnw(ctx, "cannot-load-service-alias", log.Fields{
			"err":      err,
			"fallback": alias,
		})

		//Store the fallback alias to avoid the fallback on future requests
		err := t.StoreServiceAlias(ctx, *alias)
		if err != nil {
			return nil, fmt.Errorf("cannot-store-fallback-service-alias")
		}
	}

	return alias, nil
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

	portNames := getLocationsToPortsMap(ports)

	items := []YangItem{}

	for _, service := range services {
		portName, ok := portNames[service.Location]
		if !ok {
			return nil, fmt.Errorf("no-port-name-for-location: %s", service.Location)
		}

		alias, err := t.getServiceAliasOrFallback(ctx, service.TagInfo.ServiceName, ServiceKey{
			Port: portName,
			STag: strconv.Itoa(service.TagInfo.PonSTag),
			CTag: strconv.Itoa(service.TagInfo.PonCTag),
			TpId: strconv.Itoa(service.TagInfo.TechnologyProfileID),
		})
		if err != nil {
			return nil, err
		}

		vlansItems, err := translateVlans(service.TagInfo, *alias)
		if err != nil {
			return nil, fmt.Errorf("cannot-translate-vlans: %v", err)
		}

		items = append(items, vlansItems...)
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

	portNames := getLocationsToPortsMap(ports)

	items := []YangItem{}

	for _, service := range services {
		portName, ok := portNames[service.Location]
		if !ok {
			return nil, fmt.Errorf("no-port-name-for-location: %s", service.Location)
		}

		alias, err := t.getServiceAliasOrFallback(ctx, service.TagInfo.ServiceName, ServiceKey{
			Port: portName,
			STag: strconv.Itoa(service.TagInfo.PonSTag),
			CTag: strconv.Itoa(service.TagInfo.PonCTag),
			TpId: strconv.Itoa(service.TagInfo.TechnologyProfileID),
		})
		if err != nil {
			return nil, err
		}

		serviceItems, err := translateService(service.TagInfo, *alias)
		if err != nil {
			return nil, fmt.Errorf("cannot-translate-service: %v", err)
		}

		items = append(items, serviceItems...)
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

//Used to uniquely identify the service and
//construct a KV Store path to the service info
type ServiceKey struct {
	Port string `json:"port"`
	STag string `json:"sTag"`
	CTag string `json:"cTag"`
	TpId string `json:"tpId"`
}

//Holds user provided names for the definition
//of a service in the yang datastore
type ServiceAlias struct {
	Key         ServiceKey `json:"key"`
	ServiceName string     `json:"serviceName"`
	VlansName   string     `json:"vlansName"`
}

func getServiceAliasKVPath(key ServiceKey) string {
	return fmt.Sprintf("%s/%s/%s/%s/%s", kvStoreServices, key.Port, key.STag, key.CTag, key.TpId)
}

func (t *VolthaYangAdapter) StoreServiceAlias(ctx context.Context, alias ServiceAlias) error {
	json, err := json.Marshal(alias)
	if err != nil {
		return err
	}

	if err = t.kvStore.Put(ctx, getServiceAliasKVPath(alias.Key), json); err != nil {
		return err
	}
	return nil
}

func (t *VolthaYangAdapter) LoadServiceAlias(ctx context.Context, key ServiceKey) (*ServiceAlias, error) {
	found, err := t.kvStore.Get(ctx, getServiceAliasKVPath(key))
	if err != nil {
		return nil, err
	}

	if found == nil {
		return nil, fmt.Errorf("service-alias-not-found-in-kvstore: %s", key)
	}

	var foundAlias ServiceAlias
	value, err := kvstore.ToByte(found.Value)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(value, &foundAlias); err != nil {
		return nil, err
	}

	return &foundAlias, nil
}

func (t *VolthaYangAdapter) DeleteServiceAlias(ctx context.Context, key ServiceKey) error {
	err := t.kvStore.Delete(ctx, getServiceAliasKVPath(key))
	if err != nil {
		return err
	}

	return nil
}
