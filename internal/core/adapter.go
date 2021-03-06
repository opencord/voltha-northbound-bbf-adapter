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
		err = fmt.Errorf("get-devices-failed: %v", err)
		return nil, err
	}

	items := translateDevices(*devices)

	logger.Debugw(ctx, "get-devices-success", log.Fields{"items": items})

	return items, nil
}
