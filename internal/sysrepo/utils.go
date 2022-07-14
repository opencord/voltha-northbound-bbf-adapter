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

//srErrorMsg provides a description of a sysrepo error code
func srErrorMsg(code C.int) string {
	return C.GoString(C.sr_strerror(code))
}

//lyErrorMsg provides the last libyang error message
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

//Creates a new libyang nodes tree from a set of new paths.
//The tree must bee manually freed after its use with C.lyd_free_all or
//an equivalent function
func createYangTree(ctx context.Context, session *C.sr_session_ctx_t, items []core.YangItem) (*C.lyd_node, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("no-items")
	}

	conn := C.sr_session_get_connection(session)
	if conn == nil {
		return nil, fmt.Errorf("null-connection")
	}

	//libyang context
	ly_ctx := C.sr_acquire_context(conn)
	if ly_ctx == nil {
		return nil, fmt.Errorf("null-libyang-context")
	}
	defer C.sr_release_context(conn)

	//Create parent node
	parentPath := C.CString(items[0].Path)
	parentValue := C.CString(items[0].Value)

	var parent *C.lyd_node
	lyErr := C.lyd_new_path(nil, ly_ctx, parentPath, parentValue, 0, &parent)
	if lyErr != C.LY_SUCCESS {
		err := fmt.Errorf("libyang-new-path-failed: %d %s", lyErr, lyErrorMsg(ly_ctx))
		return nil, err
	}
	logger.Debugw(ctx, "creating-yang-item", log.Fields{"item": items[0]})

	freeCString(parentPath)
	freeCString(parentValue)

	//Add remaining nodes
	for _, item := range items[1:] {
		logger.Debugw(ctx, "creating-yang-item", log.Fields{"item": item})

		path := C.CString(item.Path)
		value := C.CString(item.Value)

		lyErr := C.lyd_new_path(parent, ly_ctx, path, value, 0, nil)
		if lyErr != C.LY_SUCCESS {
			freeCString(path)
			freeCString(value)

			//Free the partially created tree
			C.lyd_free_all(parent)

			err := fmt.Errorf("libyang-new-path-failed: %d %s", lyErr, lyErrorMsg(ly_ctx))

			return nil, err
		}

		freeCString(path)
		freeCString(value)
	}

	return parent, nil
}

//Creates a set of new paths under an existing libyang tree parent node
func updateYangTree(ctx context.Context, session *C.sr_session_ctx_t, parent **C.lyd_node, items []core.YangItem) error {
	if len(items) == 0 {
		//Nothing to do
		return nil
	}

	conn := C.sr_session_get_connection(session)
	if conn == nil {
		return fmt.Errorf("null-connection")
	}

	//libyang context
	ly_ctx := C.sr_acquire_context(conn)
	if ly_ctx == nil {
		return fmt.Errorf("null-libyang-context")
	}
	defer C.sr_release_context(conn)

	for _, item := range items {
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

//Merges the content of a yang tree with the content of the datastore.
//The target datastore is the one on which the session has been created
func editDatastore(ctx context.Context, session *C.sr_session_ctx_t, editsTree *C.lyd_node) error {
	errCode := C.sr_edit_batch(session, editsTree, C.mergeOperation)
	if errCode != C.SR_ERR_OK {
		err := fmt.Errorf("failed-to-edit-datastore")
		logger.Errorw(ctx, err.Error(), log.Fields{"errCode": errCode, "errMsg": srErrorMsg(errCode)})
		return err
	}

	errCode = C.sr_apply_changes(session, 0)
	if errCode != C.SR_ERR_OK {
		err := fmt.Errorf("failed-to-apply-datastore-changes")
		logger.Errorw(ctx, err.Error(), log.Fields{"errCode": errCode, "errMsg": srErrorMsg(errCode)})
		return err
	}

	return nil
}
