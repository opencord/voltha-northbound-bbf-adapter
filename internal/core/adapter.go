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
	oltAppClient    *clients.OltAppClient
}

func NewVolthaYangAdapter(nbiClient *clients.VolthaNbiClient, oltClient *clients.OltAppClient) *VolthaYangAdapter {
	return &VolthaYangAdapter{
		volthaNbiClient: nbiClient,
		oltAppClient:    oltClient,
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
			logger.Debugw(ctx, "get-ports-success", log.Fields{"deviceId": device.Id, "ports": ports})

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
