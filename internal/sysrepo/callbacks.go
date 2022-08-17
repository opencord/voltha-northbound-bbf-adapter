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
	"strconv"

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

//export get_services_cb
func get_services_cb(session *C.sr_session_ctx_t, parent **C.lyd_node) C.sr_error_t {
	//This function is a callback for the retrieval of devices from sysrepo
	//The "export" comment instructs CGO to create a C function for it

	ctx := context.Background()
	logger.Debug(ctx, "processing-get-services-request")

	if session == nil {
		logger.Error(ctx, "sysrepo-get-services-null-session")
		return C.SR_ERR_OPERATION_FAILED
	}

	if parent == nil {
		logger.Error(ctx, "sysrepo-get-services-null-parent-node")
		return C.SR_ERR_OPERATION_FAILED
	}

	if core.AdapterInstance == nil {
		logger.Error(ctx, "sysrepo-get-services-nil-translator")
		return C.SR_ERR_OPERATION_FAILED
	}

	services, err := core.AdapterInstance.GetServices(ctx)
	if err != nil {
		logger.Errorw(ctx, "sysrepo-get-services-translation-error", log.Fields{"err": err})
		return C.SR_ERR_OPERATION_FAILED
	}

	err = updateYangTree(ctx, session, parent, services)
	if err != nil {
		logger.Errorw(ctx, "sysrepo-get-services-update-error", log.Fields{"err": err})
		return C.SR_ERR_OPERATION_FAILED
	}

	logger.Info(ctx, "services-information-request-served")

	return C.SR_ERR_OK
}

//export get_vlans_cb
func get_vlans_cb(session *C.sr_session_ctx_t, parent **C.lyd_node) C.sr_error_t {
	//This function is a callback for the retrieval of vlans from sysrepo
	//The "export" comment instructs CGO to create a C function for it

	ctx := context.Background()
	logger.Debug(ctx, "processing-get-vlans-request")

	if session == nil {
		logger.Error(ctx, "sysrepo-get-vlans-null-session")
		return C.SR_ERR_OPERATION_FAILED
	}

	if parent == nil {
		logger.Error(ctx, "sysrepo-get-vlans-null-parent-node")
		return C.SR_ERR_OPERATION_FAILED
	}

	if core.AdapterInstance == nil {
		logger.Error(ctx, "sysrepo-get-vlans-nil-translator")
		return C.SR_ERR_OPERATION_FAILED
	}

	vlans, err := core.AdapterInstance.GetVlans(ctx)
	if err != nil {
		logger.Errorw(ctx, "sysrepo-get-vlans-translation-error", log.Fields{"err": err})
		return C.SR_ERR_OPERATION_FAILED
	}

	err = updateYangTree(ctx, session, parent, vlans)
	if err != nil {
		logger.Errorw(ctx, "sysrepo-get-vlans-update-error", log.Fields{"err": err})
		return C.SR_ERR_OPERATION_FAILED
	}

	logger.Info(ctx, "vlans-information-request-served")

	return C.SR_ERR_OK
}

//export get_bandwidth_profiles_cb
func get_bandwidth_profiles_cb(session *C.sr_session_ctx_t, parent **C.lyd_node) C.sr_error_t {
	//This function is a callback for the retrieval of bandwidth profiles from sysrepo
	//The "export" comment instructs CGO to create a C function for it

	ctx := context.Background()
	logger.Debug(ctx, "processing-get-bandwidth-profiles-request")

	if session == nil {
		logger.Error(ctx, "sysrepo-get-bandwidth-profiles-null-session")
		return C.SR_ERR_OPERATION_FAILED
	}

	if parent == nil {
		logger.Error(ctx, "sysrepo-get-bandwidth-profiles-null-parent-node")
		return C.SR_ERR_OPERATION_FAILED
	}

	if core.AdapterInstance == nil {
		logger.Error(ctx, "sysrepo-get-bandwidth-profiles-nil-translator")
		return C.SR_ERR_OPERATION_FAILED
	}

	bwProfiles, err := core.AdapterInstance.GetBandwidthProfiles(ctx)
	if err != nil {
		logger.Errorw(ctx, "sysrepo-get-bandwidth-profiles-translation-error", log.Fields{"err": err})
		return C.SR_ERR_OPERATION_FAILED
	}

	err = updateYangTree(ctx, session, parent, bwProfiles)
	if err != nil {
		logger.Errorw(ctx, "sysrepo-get-bandwidth-profiles-update-error", log.Fields{"err": err})
		return C.SR_ERR_OPERATION_FAILED
	}

	logger.Info(ctx, "bandwidth-profiles-information-request-served")

	return C.SR_ERR_OK
}

//export edit_service_profiles_cb
func edit_service_profiles_cb(editSession *C.sr_session_ctx_t, runningSession *C.sr_session_ctx_t, event C.sr_event_t) C.sr_error_t {
	//This function is a callback for changes on service profiles
	//The "export" comment instructs CGO to create a C function for it

	if event != C.SR_EV_CHANGE {
		return C.SR_ERR_OK
	}

	ctx := context.Background()
	logger.Debug(ctx, "processing-service-profile-changes")

	serviceNamesChanges, err := getChangesList(ctx, editSession, core.ServiceProfilesPath+"/service-profile/name")
	if err != nil {
		logger.Errorw(ctx, "cannot-get-service-profile-names-changes", log.Fields{"err": err})
		return C.SR_ERR_OPERATION_FAILED
	}

	for _, n := range serviceNamesChanges {
		switch n.Operation {
		case C.SR_OP_CREATED:
			if errCode := edit_service_create(ctx, editSession, runningSession, n.Value); errCode != C.SR_ERR_OK {
				return errCode
			}
		case C.SR_OP_DELETED:
			if errCode := edit_service_delete(ctx, editSession, runningSession, n.Value); errCode != C.SR_ERR_OK {
				return errCode
			}
		default:
			return C.SR_ERR_UNSUPPORTED
		}
	}

	return C.SR_ERR_OK
}

func edit_service_create(ctx context.Context, editSession *C.sr_session_ctx_t, runningSession *C.sr_session_ctx_t, serviceName string) C.sr_error_t {
	portName, err := getSingleChangeValue(ctx, editSession, fmt.Sprintf("%s/service-profile[name='%s']/ports/port/name", core.ServiceProfilesPath, serviceName))
	if err != nil {
		logger.Errorw(ctx, "cannot-get-service-profile-port-changes", log.Fields{"err": err, "service": serviceName})
		return C.SR_ERR_OPERATION_FAILED
	}

	servicePortPath := core.GetServicePortPath(serviceName, portName)

	tpId, err := getSingleChangeValue(ctx, editSession, servicePortPath+"/bbf-nt-service-profile-voltha:technology-profile-id")
	if err != nil {
		logger.Errorw(ctx, "cannot-get-service-profile-tp-id-change", log.Fields{"err": err, "service": serviceName})
		return C.SR_ERR_OPERATION_FAILED
	}

	vlanName, err := getSingleChangeValue(ctx, editSession, servicePortPath+"/port-vlans/port-vlan/name")
	if err != nil {
		logger.Errorw(ctx, "cannot-get-service-profile-vlan-change", log.Fields{"err": err, "service": serviceName})
		return C.SR_ERR_OPERATION_FAILED
	}

	vlansPath := core.GetVlansPath(vlanName)

	sTag, err := getSingleChangeValue(ctx, editSession, vlansPath+"/ingress-rewrite/push-outer-tag/vlan-id")
	if err != nil {
		logger.Errorw(ctx, "cannot-get-service-profile-stag-changes", log.Fields{"err": err, "service": serviceName})
		return C.SR_ERR_OPERATION_FAILED
	}
	if sTag == core.YangVlanIdAny {
		sTag = strconv.Itoa(core.VolthaVlanIdAny)
	}

	cTag, err := getSingleChangeValue(ctx, editSession, vlansPath+"/ingress-rewrite/push-second-tag/vlan-id")
	if err != nil {
		logger.Errorw(ctx, "cannot-get-service-profile-stag-changes", log.Fields{"err": err, "service": serviceName})
		return C.SR_ERR_OPERATION_FAILED
	}
	if cTag == core.YangVlanIdAny {
		cTag = strconv.Itoa(core.VolthaVlanIdAny)
	}

	alias := core.ServiceAlias{
		Key: core.ServiceKey{
			Port: portName,
			CTag: cTag,
			STag: sTag,
			TpId: tpId,
		},
		ServiceName: serviceName,
		VlansName:   vlanName,
	}
	logger.Infow(ctx, "new-service-profile-information", log.Fields{
		"serviceInfo": alias,
	})

	if core.AdapterInstance == nil {
		logger.Error(ctx, "sysrepo-service-changes-nil-translator")
		return C.SR_ERR_OPERATION_FAILED
	}

	if err := core.AdapterInstance.ProvisionService(portName, sTag, cTag, tpId); err != nil {
		logger.Errorw(ctx, "service-provisioning-error", log.Fields{
			"service": serviceName,
			"err":     err,
		})
		return C.SR_ERR_OPERATION_FAILED
	}

	if err := core.AdapterInstance.StoreServiceAlias(ctx, alias); err != nil {
		//Log the error but don't make the callback fail
		//The service in ONOS has been provisioned succesfully and the datastore has to stay aligned
		//A fallback alias will be created if service data is requested later
		logger.Errorw(ctx, "cannot-store-service-alias-in-kvstore", log.Fields{"err": err, "service": serviceName})
	}

	logger.Infow(ctx, "service-profile-creation-request-served", log.Fields{
		"service": serviceName,
	})

	return C.SR_ERR_OK
}

func edit_service_delete(ctx context.Context, editSession *C.sr_session_ctx_t, runningSession *C.sr_session_ctx_t, serviceName string) C.sr_error_t {
	portName, err := getDatastoreLeafValue(ctx, runningSession, fmt.Sprintf("%s/service-profile[name='%s']/ports/port/name", core.ServiceProfilesPath, serviceName))
	if err != nil {
		logger.Errorw(ctx, "cannot-get-service-profile-port-leaf", log.Fields{"err": err, "service": serviceName})
		return C.SR_ERR_OPERATION_FAILED
	}

	servicePortPath := core.GetServicePortPath(serviceName, portName)

	tpId, err := getDatastoreLeafValue(ctx, runningSession, servicePortPath+"/bbf-nt-service-profile-voltha:technology-profile-id")
	if err != nil {
		logger.Errorw(ctx, "cannot-get-service-profile-tp-id-leaf", log.Fields{"err": err, "service": serviceName})
		return C.SR_ERR_OPERATION_FAILED
	}

	vlanName, err := getDatastoreLeafValue(ctx, runningSession, servicePortPath+"/port-vlans/port-vlan/name")
	if err != nil {
		logger.Errorw(ctx, "cannot-get-service-profile-vlan-leaf", log.Fields{"err": err, "service": serviceName})
		return C.SR_ERR_OPERATION_FAILED
	}

	vlansPath := core.GetVlansPath(vlanName)

	sTag, err := getDatastoreLeafValue(ctx, runningSession, vlansPath+"/ingress-rewrite/push-outer-tag/vlan-id")
	if err != nil {
		logger.Errorw(ctx, "cannot-get-service-profile-stag-leaf", log.Fields{"err": err, "service": serviceName})
		return C.SR_ERR_OPERATION_FAILED
	}
	if sTag == core.YangVlanIdAny {
		sTag = strconv.Itoa(core.VolthaVlanIdAny)
	}

	cTag, err := getDatastoreLeafValue(ctx, runningSession, vlansPath+"/ingress-rewrite/push-second-tag/vlan-id")
	if err != nil {
		logger.Errorw(ctx, "cannot-get-service-profile-stag-leaf", log.Fields{"err": err, "service": serviceName})
		return C.SR_ERR_OPERATION_FAILED
	}
	if cTag == core.YangVlanIdAny {
		cTag = strconv.Itoa(core.VolthaVlanIdAny)
	}

	alias := core.ServiceAlias{
		Key: core.ServiceKey{
			Port: portName,
			CTag: cTag,
			STag: sTag,
			TpId: tpId,
		},
		ServiceName: serviceName,
		VlansName:   vlanName,
	}
	logger.Infow(ctx, "service-profile-deletion-information", log.Fields{
		"serviceInfo": alias,
	})

	if err := core.AdapterInstance.RemoveService(portName, sTag, cTag, tpId); err != nil {
		logger.Errorw(ctx, "service-removal-error", log.Fields{
			"service": serviceName,
			"err":     err,
		})
		return C.SR_ERR_OPERATION_FAILED
	}

	if err := core.AdapterInstance.DeleteServiceAlias(ctx, alias.Key); err != nil {
		//Log the error but don't make the callback fail
		//The service in ONOS has been removed succesfully and the datastore has to stay aligned
		//The only side effect is a dangling alias left in the KV store
		logger.Errorw(ctx, "cannot-delete-service-alias-from-kvstore", log.Fields{"err": err, "service": serviceName})
	}

	logger.Infow(ctx, "service-profile-removal-request-served", log.Fields{
		"service": serviceName,
	})

	return C.SR_ERR_OK
}

//export edit_vlans_cb
func edit_vlans_cb(editSession *C.sr_session_ctx_t, event C.sr_event_t) C.sr_error_t {
	//This function is a callback for changes on VLANs
	//The "export" comment instructs CGO to create a C function for it

	if event != C.SR_EV_CHANGE {
		return C.SR_ERR_OK
	}

	ctx := context.Background()
	logger.Debug(ctx, "processing-vlans-changes")

	vlanChanges, err := getChangesList(ctx, editSession, core.VlansPath+"//.")
	if err != nil {
		logger.Errorw(ctx, "cannot-get-vlans-changes", log.Fields{"err": err})
		return C.SR_ERR_OPERATION_FAILED
	}

	for _, n := range vlanChanges {
		//VLANs must be defined through creation (for service provisioning)
		//or deletion (for service removal). Changes to the VLAN values
		//are not supported, because VOLTHA does not support dynamic changes
		//to the service.
		switch n.Operation {
		case C.SR_OP_CREATED:
		case C.SR_OP_DELETED:
			//Everything will be handled in the services callback
			//Just approve the change here
			return C.SR_ERR_OK
		default:
			return C.SR_ERR_UNSUPPORTED
		}
	}

	return C.SR_ERR_OK
}
