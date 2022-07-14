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

	"github.com/opencord/voltha-lib-go/v7/pkg/log"
	"github.com/opencord/voltha-northbound-bbf-adapter/internal/core"
	"github.com/opencord/voltha-protos/v5/go/voltha"
)

const (
	eventNameOnuActivated = "ONU_ACTIVATED_RAISE_EVENT"
)

//Performs the necessary operations on a new voltha event received from Kafka
func (p *SysrepoPlugin) ManageVolthaEvent(ctx context.Context, event *voltha.Event) {
	if event.Header.Type == voltha.EventType_DEVICE_EVENT {
		devEvent, ok := event.EventType.(*voltha.Event_DeviceEvent)
		if !ok {
			logger.Errorw(ctx, "unexpected-event-type", log.Fields{
				"headerType": event.Header.Type,
				"actualType": fmt.Sprintf("%T", event.EventType),
			})
			return
		}

		//TODO: map other events to ONU state changes
		switch devEvent.DeviceEvent.DeviceEventName {
		case eventNameOnuActivated:
			logger.Debugw(ctx, "onu-activated-event-received", log.Fields{
				"header":      event.Header,
				"deviceEvent": devEvent.DeviceEvent,
			})

			if err := p.sendOnuActivatedNotification(ctx, event.Header, devEvent.DeviceEvent); err != nil {
				logger.Errorw(ctx, "failed-to-send-onu-activated-notification", log.Fields{"err": err})
			}
		}
	}
}

//Sends a notification based on the content of the received device event
func (p *SysrepoPlugin) sendOnuActivatedNotification(ctx context.Context, eventHeader *voltha.EventHeader, deviceEvent *voltha.DeviceEvent) error {
	//Prepare the content of the notification
	notificationItems, channelTermItems, err := core.TranslateOnuActivatedEvent(eventHeader, deviceEvent)
	if err != nil {
		return fmt.Errorf("failed-to-translate-onu-activated-event: %v", err)
	}

	//Create the channel termination in the datastore to make the notification leafref valid
	channelTermTree, err := createYangTree(ctx, p.operationalSession, channelTermItems)
	if err != nil {
		return fmt.Errorf("failed-to-create-channel-termination-tree: %v", err)
	}
	defer C.lyd_free_all(channelTermTree)

	err = editDatastore(ctx, p.operationalSession, channelTermTree)
	if err != nil {
		return fmt.Errorf("failed-to-apply-channel-termination-to-datastore: %v", err)
	}

	//Create the notification tree
	notificationTree, err := createYangTree(ctx, p.operationalSession, notificationItems)
	if err != nil {
		return fmt.Errorf("failed-to-create-onu-activated-notification-tree: %v", err)
	}

	//Let sysrepo manage the notification tree to properly free it after its delivery
	var notificationData *C.sr_data_t
	errCode := C.sr_acquire_data(p.connection, notificationTree, &notificationData)
	if errCode != C.SR_ERR_OK {
		err := fmt.Errorf("cannot-acquire-notification-data")
		logger.Errorw(ctx, err.Error(), log.Fields{"errCode": errCode, "errMsg": srErrorMsg(errCode)})
		return err
	}
	defer C.sr_release_data(notificationData)

	//Send the notification
	logger.Infow(ctx, "sending-onu-activated-notification", log.Fields{
		"onuSn": deviceEvent.Context["serial-number"],
	})
	errCode = C.sr_notif_send_tree(p.operationalSession, notificationData.tree, 0, 0)
	if errCode != C.SR_ERR_OK {
		err := fmt.Errorf("cannot-send-notification")
		logger.Errorw(ctx, err.Error(), log.Fields{"errCode": errCode, "errMsg": srErrorMsg(errCode)})
		return err
	}

	return nil
}
