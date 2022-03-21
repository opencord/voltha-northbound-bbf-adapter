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

//Exported by sysrepo.go
sr_error_t get_devices_cb(sr_session_ctx_t *session, lyd_node **parent);

//The wrapper function is needed because CGO cannot express const char*
//and thus it can't match sysrepo's callback signature
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