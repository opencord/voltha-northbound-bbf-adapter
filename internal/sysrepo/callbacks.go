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

package sysrepo

//#cgo LDFLAGS: -lsysrepo -lyang -Wl,--allow-multiple-definition
//#include "plugin.c"
import "C"
import (
	"context"

	"github.com/opencord/voltha-lib-go/v7/pkg/log"
	"github.com/opencord/voltha-northbound-bbf-adapter/internal/core"
)

//export get_devices_cb
func get_devices_cb(session *C.sr_session_ctx_t, parent **C.lyd_node) C.sr_error_t {
	//This function is a callback for the retrieval of devices from sysrepo
	//The "export" comment instructs CGO to create a C function for it

	ctx := context.Background()
	logger.Debug(ctx, "processing-get-devices-request")

	if session == nil {
		logger.Error(ctx, "sysrepo-get-devices-null-session")
		return C.SR_ERR_OPERATION_FAILED
	}

	if parent == nil {
		logger.Error(ctx, "sysrepo-get-devices-null-parent-node")
		return C.SR_ERR_OPERATION_FAILED
	}

	if core.AdapterInstance == nil {
		logger.Error(ctx, "sysrepo-get-devices-nil-translator")
		return C.SR_ERR_OPERATION_FAILED
	}

	devices, err := core.AdapterInstance.GetDevices(ctx)
	if err != nil {
		logger.Errorw(ctx, "sysrepo-get-devices-translator-error", log.Fields{"err": err})
		return C.SR_ERR_OPERATION_FAILED
	}

	err = updateYangTree(ctx, session, parent, devices)
	if err != nil {
		logger.Errorw(ctx, "sysrepo-get-devices-update-error", log.Fields{"err": err})
		return C.SR_ERR_OPERATION_FAILED
	}

	logger.Info(ctx, "devices-information-request-served")

	return C.SR_ERR_OK
}
