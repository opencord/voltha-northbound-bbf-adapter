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

#include <sysrepo.h>
#include <sysrepo/xpath.h>

//Needed to handle callback functions with a working data type in CGO
typedef void (*function)(); // https://golang.org/issue/19835

//Exported by sysrepo.go
void get_data_cb();

int get_data_cb_wrapper(
    sr_session_ctx_t *session,
    uint32_t subscription_id,
    const char *module_name,
    const char *path,
    const char *request_xpath,
    uint32_t request_id,
    struct lyd_node **parent,
    void *private_data)
{
    get_data_cb();

    return SR_ERR_OK;
}