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

package clients

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/opencord/voltha-lib-go/v7/pkg/log"
)

const (
	oltAppHttpRequestTimeout = time.Second * 10
	oltAppBackoffInterval    = time.Second * 10
)

type OltAppClient struct {
	httpClient *http.Client
	endpoint   string
	username   string
	password   string
}

type RestResponse struct {
	Body string
	Code int
}

// Creates a new olt app client
func NewOltAppClient(endpoint string, user string, pass string) *OltAppClient {
	return &OltAppClient{
		httpClient: &http.Client{
			Timeout: oltAppHttpRequestTimeout,
		},
		endpoint: endpoint,
		username: user,
		password: pass,
	}
}

func (c *OltAppClient) CheckConnection(ctx context.Context) error {
	logger.Debugw(ctx, "checking-connection-to-onos-olt-app-api", log.Fields{"endpoint": c.endpoint})

	for {
		if resp, err := c.GetStatus(); err == nil {
			logger.Debug(ctx, "onos-olt-app-api-reachable")
			break
		} else {
			logger.Warnw(ctx, "onos-olt-app-api-not-ready", log.Fields{
				"err":      err,
				"response": resp,
			})
		}

		//Wait a bit before trying again
		select {
		case <-ctx.Done():
			return fmt.Errorf("onos-olt-app-connection-stopped-due-to-context-done")
		case <-time.After(oltAppBackoffInterval):
			continue
		}
	}

	return nil
}

func (c *OltAppClient) makeRequest(method string, url string) (RestResponse, error) {
	result := RestResponse{Code: 0}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return result, fmt.Errorf("cannot-create-request: %s", err)
	}

	req.SetBasicAuth(c.username, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return result, fmt.Errorf("cannot-get-response: %s", err)
	}
	defer resp.Body.Close()

	buffer, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, fmt.Errorf("error-while-reading-response-body: %s", err)
	}

	result.Body = string(buffer)
	result.Code = resp.StatusCode

	if result.Code != http.StatusOK {
		return result, fmt.Errorf("status-code-not-ok: %s %s %d", method, url, result.Code)
	}

	return result, nil
}

func (c *OltAppClient) GetStatus() (RestResponse, error) {
	method := http.MethodGet
	url := fmt.Sprintf("http://%s/onos/olt/oltapp/status", c.endpoint)

	return c.makeRequest(method, url)
}

//NOTE: if methods are used to retrieve more complex information
//it may be better to return an already deserialized structure
//instead of the current RestResponse
func (c *OltAppClient) ProvisionSubscriber(device string, port uint32) (RestResponse, error) {
	method := http.MethodPost
	url := fmt.Sprintf("http://%s/onos/olt/oltapp/%s/%d", c.endpoint, device, port)

	return c.makeRequest(method, url)
}

func (c *OltAppClient) RemoveSubscriber(device string, port uint32) (RestResponse, error) {
	method := http.MethodDelete
	url := fmt.Sprintf("http://%s/onos/olt/oltapp/%s/%d", c.endpoint, device, port)

	return c.makeRequest(method, url)
}
