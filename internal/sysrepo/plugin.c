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

#include <libyang/libyang.h>
#include <sysrepo.h>
#include <sysrepo/xpath.h>

//Needed to handle callback functions with a working data type in CGO
typedef void (*function)(); // https://golang.org/issue/19835

//CGO can't see raw structs
typedef struct lyd_node lyd_node;
typedef struct ly_ctx ly_ctx;

//Used to define the datastore edit mode
const char* mergeOperation = "merge";

//Provides data for the schema-mount extension
LY_ERR mountpoint_ext_data_clb(
    const struct lysc_ext_instance *ext,
    void *user_data,
    void **ext_data,
    ly_bool *ext_data_free)
{
    *ext_data = (lyd_node*) user_data;
    *ext_data_free = 0;
    return LY_SUCCESS;
}

// Exported by callbacks.go
sr_error_t get_devices_cb(sr_session_ctx_t *session, lyd_node **parent);

//The wrapper functions are needed because CGO cannot express some keywords
//such as "const", and thus it can't match sysrepo's callback signature

int get_devices_cb_wrapper(
    sr_session_ctx_t *session,
    uint32_t subscription_id,
    const char *module_name,
    const char *path,
    const char *request_xpath,
    uint32_t request_id,
    struct lyd_node **parent,
    void *private_data)
{
    return get_devices_cb(session, parent);
}