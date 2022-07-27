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
	connection         *C.sr_conn_ctx_t
	operationalSession *C.sr_session_ctx_t
	runningSession     *C.sr_session_ctx_t
	subscription       *C.sr_subscription_ctx_t
	schemaMountData    *C.lyd_node
}

//createPluginState populates a SysrepoPlugin struct by establishing
//a connection and a session
func (p *SysrepoPlugin) createSessions(ctx context.Context) error {
	var errCode C.int

	//Populates connection
	errCode = C.sr_connect(C.SR_CONN_DEFAULT, &p.connection)
	if errCode != C.SR_ERR_OK {
		err := fmt.Errorf("sysrepo-connect-error")
		logger.Errorw(ctx, err.Error(), log.Fields{"errCode": errCode, "errMsg": srErrorMsg(errCode)})
		return err
	}

	//Populates sessions
	//The session on the operation datastore will be used for most operations
	//The session on the running datastore will be used for the subscription to edits
	//since the operational datastore can't be edited by the client
	errCode = C.sr_session_start(p.connection, C.SR_DS_OPERATIONAL, &p.operationalSession)
	if errCode != C.SR_ERR_OK {
		err := fmt.Errorf("sysrepo-operational-session-error")
		logger.Errorw(ctx, err.Error(), log.Fields{"errCode": errCode, "errMsg": srErrorMsg(errCode)})

		_ = p.Stop(ctx)

		return err
	}

	errCode = C.sr_session_start(p.connection, C.SR_DS_RUNNING, &p.runningSession)
	if errCode != C.SR_ERR_OK {
		err := fmt.Errorf("sysrepo-running-session-error")
		logger.Errorw(ctx, err.Error(), log.Fields{"errCode": errCode, "errMsg": srErrorMsg(errCode)})

		_ = p.Stop(ctx)

		return err
	}

	return nil
}

func StartNewPlugin(ctx context.Context, schemaMountFilePath string) (*SysrepoPlugin, error) {
	plugin := &SysrepoPlugin{}

	//Set sysrepo and libyang log level
	if logger.GetLogLevel() == log.DebugLevel {
		C.sr_log_stderr(C.SR_LL_INF)
		C.ly_log_level(C.LY_LLVRB)
	} else {
		C.sr_log_stderr(C.SR_LL_ERR)
		C.ly_log_level(C.LY_LLERR)
	}

	//Open a session to sysrepo
	err := plugin.createSessions(ctx)
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
	devicesModule := C.CString(core.DeviceAggregationModule)
	devicesPath := C.CString(core.DevicesPath + "/*")
	defer freeCString(devicesModule)
	defer freeCString(devicesPath)

	servicesModule := C.CString(core.ServiceProfileModule)
	servicesPath := C.CString(core.ServiceProfilesPath + "/*")
	defer freeCString(servicesModule)
	defer freeCString(servicesPath)

	vlansModule := C.CString(core.VlansModule)
	vlansPath := C.CString(core.VlansPath + "/*")
	defer freeCString(vlansModule)
	defer freeCString(vlansPath)

	bwProfilesModule := C.CString(core.BandwidthProfileModule)
	bwProfilesPath := C.CString(core.BandwidthProfilesPath + "/*")
	defer freeCString(bwProfilesModule)
	defer freeCString(bwProfilesPath)

	//Get devices
	errCode := C.sr_oper_get_subscribe(
		plugin.operationalSession,
		devicesModule,
		devicesPath,
		C.function(C.get_devices_cb_wrapper),
		C.NULL,
		C.SR_SUBSCR_DEFAULT,
		&plugin.subscription,
	)
	if errCode != C.SR_ERR_OK {
		err := fmt.Errorf("sysrepo-failed-subscription-to-get-devices")
		logger.Errorw(ctx, err.Error(), log.Fields{"errCode": errCode, "errMsg": srErrorMsg(errCode)})
		return nil, err
	}

	//Get services
	errCode = C.sr_oper_get_subscribe(
		plugin.operationalSession,
		servicesModule,
		servicesPath,
		C.function(C.get_services_cb_wrapper),
		C.NULL,
		C.SR_SUBSCR_DEFAULT,
		&plugin.subscription,
	)
	if errCode != C.SR_ERR_OK {
		err := fmt.Errorf("sysrepo-failed-subscription-to-get-services")
		logger.Errorw(ctx, err.Error(), log.Fields{"errCode": errCode, "errMsg": srErrorMsg(errCode)})
		return nil, err
	}

	//Get vlans
	errCode = C.sr_oper_get_subscribe(
		plugin.operationalSession,
		vlansModule,
		vlansPath,
		C.function(C.get_vlans_cb_wrapper),
		C.NULL,
		C.SR_SUBSCR_DEFAULT,
		&plugin.subscription,
	)
	if errCode != C.SR_ERR_OK {
		err := fmt.Errorf("sysrepo-failed-subscription-to-get-services")
		logger.Errorw(ctx, err.Error(), log.Fields{"errCode": errCode, "errMsg": srErrorMsg(errCode)})
		return nil, err
	}

	//Get bandwidth profiles
	errCode = C.sr_oper_get_subscribe(
		plugin.operationalSession,
		bwProfilesModule,
		bwProfilesPath,
		C.function(C.get_bandwidth_profiles_cb_wrapper),
		C.NULL,
		C.SR_SUBSCR_DEFAULT,
		&plugin.subscription,
	)
	if errCode != C.SR_ERR_OK {
		err := fmt.Errorf("sysrepo-failed-subscription-to-get-services")
		logger.Errorw(ctx, err.Error(), log.Fields{"errCode": errCode, "errMsg": srErrorMsg(errCode)})
		return nil, err
	}

	//Subscribe with a callback to changes of configuration in the services modules
	//Changes to services
	errCode = C.sr_module_change_subscribe(
		plugin.runningSession,
		servicesModule,
		servicesPath,
		C.function(C.edit_service_profiles_cb_wrapper),
		unsafe.Pointer(plugin.runningSession), //Pass session for running datastore to get current data
		0,
		C.SR_SUBSCR_DEFAULT,
		&plugin.subscription,
	)
	if errCode != C.SR_ERR_OK {
		err := fmt.Errorf("sysrepo-failed-subscription-to-change-services")
		logger.Errorw(ctx, err.Error(), log.Fields{"errCode": errCode, "errMsg": srErrorMsg(errCode)})
		return nil, err
	}

	//Changes to VLANs
	errCode = C.sr_module_change_subscribe(
		plugin.runningSession,
		vlansModule,
		vlansPath,
		C.function(C.edit_vlans_cb_wrapper),
		C.NULL,
		0,
		C.SR_SUBSCR_DEFAULT,
		&plugin.subscription,
	)
	if errCode != C.SR_ERR_OK {
		err := fmt.Errorf("sysrepo-failed-subscription-to-change-vlans")
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

	//Frees sessions
	if p.operationalSession != nil {
		errCode = C.sr_session_stop(p.operationalSession)
		if errCode != C.SR_ERR_OK {
			err := fmt.Errorf("failed-to-close-operational-session")
			logger.Errorw(ctx, err.Error(), log.Fields{"errCode": errCode, "errMsg": srErrorMsg(errCode)})
			return err
		}
		p.operationalSession = nil
	}

	if p.runningSession != nil {
		errCode = C.sr_session_stop(p.runningSession)
		if errCode != C.SR_ERR_OK {
			err := fmt.Errorf("failed-to-close-running-session")
			logger.Errorw(ctx, err.Error(), log.Fields{"errCode": errCode, "errMsg": srErrorMsg(errCode)})
			return err
		}
		p.runningSession = nil
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
