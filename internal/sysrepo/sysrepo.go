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

//#cgo CFLAGS: -I/usr/include
//#cgo LDFLAGS: -lsysrepo -Wl,--allow-multiple-definition
//#include "plugin.c"
import "C"
import (
	"context"
	"fmt"

	"github.com/opencord/voltha-lib-go/v7/pkg/log"
)

const (
	BASE_YANG_MODEL    = "bbf-device-aggregation"
	DEVICES_YANG_MODEL = "/" + BASE_YANG_MODEL + ":devices"
)

type SysrepoPlugin struct {
	connection   *C.sr_conn_ctx_t
	session      *C.sr_session_ctx_t
	subscription *C.sr_subscription_ctx_t
}

func errorMsg(code C.int) string {
	return C.GoString(C.sr_strerror(code))
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

//export get_data_cb
func get_data_cb() {
	//This function is a callback for the retrieval of data from sysrepo
	//The "export" comment instructs CGO to create a C function for it

	//As a placeholder, it just reports that a request to get data
	//has been received from the netconf server

	//TODO: get actual information
	ctx := context.Background()
	logger.Info(ctx, ">>>>>>>RECEIVED REQUEST FROM SYSREPO<<<<<<<")
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
	errCode := C.sr_oper_get_items_subscribe(
		plugin.session,
		C.CString(BASE_YANG_MODEL),
		C.CString(DEVICES_YANG_MODEL+"/*"),
		C.function(C.get_data_cb_wrapper),
		C.NULL,
		C.SR_SUBSCR_CTX_REUSE,
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
