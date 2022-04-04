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
	"io/ioutil"
	"os"
	"unsafe"

	"github.com/opencord/voltha-lib-go/v7/pkg/log"
	"github.com/opencord/voltha-northbound-bbf-adapter/internal/core"
)

type SysrepoPlugin struct {
	connection      *C.sr_conn_ctx_t
	session         *C.sr_session_ctx_t
	subscription    *C.sr_subscription_ctx_t
	schemaMountData *C.lyd_node
}

func srErrorMsg(code C.int) string {
	return C.GoString(C.sr_strerror(code))
}

func lyErrorMsg(ly_ctx *C.ly_ctx) string {
	lyErrString := C.ly_errmsg(ly_ctx)
	defer freeCString(lyErrString)

	return C.GoString(lyErrString)
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
	defer C.sr_release_context(conn)
	if ly_ctx == nil {
		return fmt.Errorf("null-libyang-context")
	}

	for _, item := range items {
		if item.Value == "" {
			continue
		}

		logger.Debugw(ctx, "updating-yang-item", log.Fields{"item": item})

		path := C.CString(item.Path)
		value := C.CString(item.Value)

		lyErr := C.lyd_new_path(*parent, ly_ctx, path, value, 0, nil)
		if lyErr != C.LY_SUCCESS {
			freeCString(path)
			freeCString(value)

			err := fmt.Errorf("libyang-new-path-failed: %d %s", lyErr, lyErrorMsg(ly_ctx))

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
		logger.Errorw(ctx, err.Error(), log.Fields{"errCode": errCode, "errMsg": srErrorMsg(errCode)})
		return err
	}

	//Populates session
	errCode = C.sr_session_start(p.connection, C.SR_DS_RUNNING, &p.session)
	if errCode != C.SR_ERR_OK {
		err := fmt.Errorf("sysrepo-session-error")
		logger.Errorw(ctx, err.Error(), log.Fields{"errCode": errCode, "errMsg": srErrorMsg(errCode)})

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

func StartNewPlugin(ctx context.Context, schemaMountFilePath string) (*SysrepoPlugin, error) {
	plugin := &SysrepoPlugin{}

	//Open a session to sysrepo
	err := plugin.createSession(ctx)
	if err != nil {
		return nil, err
	}

	//Read the schema-mount file
	if _, err := os.Stat(schemaMountFilePath); err != nil {
		//The file cannot be found
		return nil, fmt.Errorf("plugin-startup-schema-mount-file-not-found: %v", err)
	}

	smBuffer, err := ioutil.ReadFile(schemaMountFilePath)
	if err != nil {
		return nil, fmt.Errorf("plugin-startup-cannot-read-schema-mount-file: %v", err)
	}

	smString := C.CString(string(smBuffer))
	defer freeCString(smString)

	ly_ctx := C.sr_acquire_context(plugin.connection)
	defer C.sr_release_context(plugin.connection)
	if ly_ctx == nil {
		return nil, fmt.Errorf("plugin-startup-null-libyang-context")
	}

	//Parse the schema-mount file into libyang nodes, and save them into the plugin data
	lyErrCode := C.lyd_parse_data_mem(ly_ctx, smString, C.LYD_XML, C.LYD_PARSE_STRICT, C.LYD_VALIDATE_PRESENT, &plugin.schemaMountData)
	if lyErrCode != C.LY_SUCCESS {
		return nil, fmt.Errorf("plugin-startup-cannot-parse-schema-mount: %v", lyErrorMsg(ly_ctx))
	}

	//Bind the callback needed to support schema-mount
	C.sr_set_ext_data_cb(plugin.connection, C.function(C.mountpoint_ext_data_clb), unsafe.Pointer(plugin.schemaMountData))

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
		logger.Errorw(ctx, err.Error(), log.Fields{"errCode": errCode, "errMsg": srErrorMsg(errCode)})
		return nil, err
	}

	logger.Debug(ctx, "sysrepo-plugin-started")

	return plugin, nil
}

func (p *SysrepoPlugin) Stop(ctx context.Context) error {
	var errCode C.int

	//Free the libyang nodes for external schema-mount data
	C.lyd_free_all(p.schemaMountData)

	//Frees subscription
	if p.subscription != nil {
		errCode = C.sr_unsubscribe(p.subscription)
		if errCode != C.SR_ERR_OK {
			err := fmt.Errorf("failed-to-close-sysrepo-subscription")
			logger.Errorw(ctx, err.Error(), log.Fields{"errCode": errCode, "errMsg": srErrorMsg(errCode)})
			return err
		}
		p.subscription = nil
	}

	//Frees session
	if p.session != nil {
		errCode = C.sr_session_stop(p.session)
		if errCode != C.SR_ERR_OK {
			err := fmt.Errorf("failed-to-close-sysrepo-session")
			logger.Errorw(ctx, err.Error(), log.Fields{"errCode": errCode, "errMsg": srErrorMsg(errCode)})
			return err
		}
		p.session = nil
	}

	//Frees connection
	if p.connection != nil {
		errCode = C.sr_disconnect(p.connection)
		if errCode != C.SR_ERR_OK {
			err := fmt.Errorf("failed-to-close-sysrepo-connection")
			logger.Errorw(ctx, err.Error(), log.Fields{"errCode": errCode, "errMsg": srErrorMsg(errCode)})
			return err
		}
		p.connection = nil
	}

	logger.Debug(ctx, "sysrepo-plugin-stopped")

	return nil
}
