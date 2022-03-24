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
	"fmt"
	"unsafe"

	"github.com/opencord/voltha-lib-go/v7/pkg/log"
	"github.com/opencord/voltha-northbound-bbf-adapter/internal/core"
)

type SysrepoPlugin struct {
	connection   *C.sr_conn_ctx_t
	session      *C.sr_session_ctx_t
	subscription *C.sr_subscription_ctx_t
}

func errorMsg(code C.int) string {
	return C.GoString(C.sr_strerror(code))
}

func freeCString(str *C.char) {
	if str != nil {
		C.free(unsafe.Pointer(str))
		str = nil
	}
}

func updateYangItems(ctx context.Context, session *C.sr_session_ctx_t, parent **C.lyd_node, items []core.YangItem) error {
	conn := C.sr_session_get_connection(session)
	if conn == nil {
		return fmt.Errorf("null-connection")
	}

	//libyang context
	ly_ctx := C.sr_acquire_context(conn)
	if ly_ctx == nil {
		return fmt.Errorf("null-libyang-context")
	}

	for _, item := range items {
		logger.Debugw(ctx, "updating-yang-item", log.Fields{"item": item})

		path := C.CString(item.Path)
		value := C.CString(item.Value)

		lyErr := C.lyd_new_path(*parent, ly_ctx, path, value, 0, nil)
		if lyErr != C.LY_SUCCESS {
			freeCString(path)
			freeCString(value)

			lyErrString := C.ly_errmsg(ly_ctx)
			err := fmt.Errorf("libyang-new-path-failed: %d %s", lyErr, C.GoString(lyErrString))
			freeCString(lyErrString)

			return err
		}

		freeCString(path)
		freeCString(value)
	}

	return nil
}

//createPluginState populates a SysrepoPlugin struct by establishing
//a connection and a session
func (p *SysrepoPlugin) createSession(ctx context.Context) error {
	var errCode C.int

	//Populates connection
	errCode = C.sr_connect(C.SR_CONN_DEFAULT, &p.connection)
	if errCode != C.SR_ERR_OK {
		err := fmt.Errorf("sysrepo-connect-error")
		logger.Errorw(ctx, err.Error(), log.Fields{"errCode": errCode, "errMsg": errorMsg(errCode)})
		return err
	}

	//Populates session
	errCode = C.sr_session_start(p.connection, C.SR_DS_RUNNING, &p.session)
	if errCode != C.SR_ERR_OK {
		err := fmt.Errorf("sysrepo-session-error")
		logger.Errorw(ctx, err.Error(), log.Fields{"errCode": errCode, "errMsg": errorMsg(errCode)})

		_ = p.Stop(ctx)

		return err
	}

	return nil
}

//export get_devices_cb
func get_devices_cb(session *C.sr_session_ctx_t, parent **C.lyd_node) C.sr_error_t {
	//This function is a callback for the retrieval of devices from sysrepo
	//The "export" comment instructs CGO to create a C function for it

	ctx := context.Background()
	logger.Debug(ctx, "processing-get-data-request")

	if session == nil {
		logger.Error(ctx, "sysrepo-get-data-null-session")
		return C.SR_ERR_OPERATION_FAILED
	}

	if parent == nil {
		logger.Error(ctx, "sysrepo-get-data-null-parent-node")
		return C.SR_ERR_OPERATION_FAILED
	}

	if core.AdapterInstance == nil {
		logger.Error(ctx, "sysrepo-get-data-nil-translator")
		return C.SR_ERR_OPERATION_FAILED
	}

	devices, err := core.AdapterInstance.GetDevices(ctx)
	if err != nil {
		logger.Errorw(ctx, "sysrepo-get-data-translator-error", log.Fields{"err": err})
		return C.SR_ERR_OPERATION_FAILED
	}

	err = updateYangItems(ctx, session, parent, devices)
	if err != nil {
		logger.Errorw(ctx, "sysrepo-get-data-update-error", log.Fields{"err": err})
		return C.SR_ERR_OPERATION_FAILED
	}

	return C.SR_ERR_OK
}

func StartNewPlugin(ctx context.Context) (*SysrepoPlugin, error) {
	plugin := &SysrepoPlugin{}

	//Open a session to sysrepo
	err := plugin.createSession(ctx)
	if err != nil {
		return nil, err
	}

	//TODO: could be useful to set it according to the adapter log level
	C.sr_log_stderr(C.SR_LL_WRN)

	//Set callbacks for events

	//Subscribe with a callback to the request of data on a certain path
	module := C.CString(core.DeviceAggregationModel)
	defer freeCString(module)

	path := C.CString(core.DevicesPath + "/*")
	defer freeCString(path)

	errCode := C.sr_oper_get_subscribe(
		plugin.session,
		module,
		path,
		C.function(C.get_devices_cb_wrapper),
		C.NULL,
		C.SR_SUBSCR_DEFAULT,
		&plugin.subscription,
	)
	if errCode != C.SR_ERR_OK {
		err := fmt.Errorf("sysrepo-failed-subscription-to-get-events")
		logger.Errorw(ctx, err.Error(), log.Fields{"errCode": errCode, "errMsg": errorMsg(errCode)})
		return nil, err
	}

	logger.Debug(ctx, "sysrepo-plugin-started")

	return plugin, nil
}

func (p *SysrepoPlugin) Stop(ctx context.Context) error {
	var errCode C.int

	//Frees subscription
	if p.subscription != nil {
		errCode = C.sr_unsubscribe(p.subscription)
		if errCode != C.SR_ERR_OK {
			err := fmt.Errorf("failed-to-close-sysrepo-subscription")
			logger.Errorw(ctx, err.Error(), log.Fields{"errCode": errCode, "errMsg": errorMsg(errCode)})
			return err
		}
		p.subscription = nil
	}

	//Frees session
	if p.session != nil {
		errCode = C.sr_session_stop(p.session)
		if errCode != C.SR_ERR_OK {
			err := fmt.Errorf("failed-to-close-sysrepo-session")
			logger.Errorw(ctx, err.Error(), log.Fields{"errCode": errCode, "errMsg": errorMsg(errCode)})
			return err
		}
		p.session = nil
	}

	//Frees connection
	if p.connection != nil {
		errCode = C.sr_disconnect(p.connection)
		if errCode != C.SR_ERR_OK {
			err := fmt.Errorf("failed-to-close-sysrepo-connection")
			logger.Errorw(ctx, err.Error(), log.Fields{"errCode": errCode, "errMsg": errorMsg(errCode)})
			return err
		}
		p.connection = nil
	}

	logger.Debug(ctx, "sysrepo-plugin-stopped")

	return nil
}
