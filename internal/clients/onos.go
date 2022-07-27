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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/opencord/voltha-lib-go/v7/pkg/log"
)

const (
	onosHttpRequestTimeout = time.Second * 10
	onosBackoffInterval    = time.Second * 10
)

type OnosClient struct {
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
func NewOnosClient(endpoint string, user string, pass string) *OnosClient {
	return &OnosClient{
		httpClient: &http.Client{
			Timeout: onosHttpRequestTimeout,
		},
		endpoint: endpoint,
		username: user,
		password: pass,
	}
}

func (c *OnosClient) CheckConnection(ctx context.Context) error {
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
		case <-time.After(onosBackoffInterval):
			continue
		}
	}

	return nil
}

func (c *OnosClient) makeRequest(method string, url string) (RestResponse, error) {
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

///////////////////////////////////////////////////////////////////////// ONOS OLT app APIs

func (c *OnosClient) GetStatus() (RestResponse, error) {
	method := http.MethodGet
	url := fmt.Sprintf("http://%s/onos/olt/oltapp/status", c.endpoint)

	return c.makeRequest(method, url)
}

func (c *OnosClient) ProvisionService(portName string, sTag string, cTag string, technologyProfileId string) (RestResponse, error) {
	method := http.MethodPost
	url := fmt.Sprintf("http://%s/onos/olt/oltapp/services/%s/%s/%s/%s", c.endpoint, portName, sTag, cTag, technologyProfileId)

	return c.makeRequest(method, url)
}

func (c *OnosClient) RemoveService(portName string, sTag string, cTag string, trafficProfileId string) (RestResponse, error) {
	method := http.MethodDelete
	url := fmt.Sprintf("http://%s/onos/olt/oltapp/services/%s/%s/%s/%s", c.endpoint, portName, sTag, cTag, trafficProfileId)

	return c.makeRequest(method, url)
}

type ProgrammedSubscriber struct {
	Location string      `json:"location"`
	TagInfo  SadisUniTag `json:"tagInfo"`
}

type SadisUniTag struct {
	UniTagMatch                   int    `json:"uniTagMatch,omitempty"`
	PonCTag                       int    `json:"ponCTag,omitempty"`
	PonSTag                       int    `json:"ponSTag,omitempty"`
	TechnologyProfileID           int    `json:"technologyProfileId,omitempty"`
	UpstreamBandwidthProfile      string `json:"upstreamBandwidthProfile,omitempty"`
	UpstreamOltBandwidthProfile   string `json:"upstreamOltBandwidthProfile,omitempty"`
	DownstreamBandwidthProfile    string `json:"downstreamBandwidthProfile,omitempty"`
	DownstreamOltBandwidthProfile string `json:"downstreamOltBandwidthProfile,omitempty"`
	IsDhcpRequired                bool   `json:"isDhcpRequired,omitempty"`
	IsIgmpRequired                bool   `json:"isIgmpRequired,omitempty"`
	IsPPPoERequired               bool   `json:"isPppoeRequired,omitempty"`
	ConfiguredMacAddress          string `json:"configuredMacAddress,omitempty"`
	EnableMacLearning             bool   `json:"enableMacLearning,omitempty"`
	UsPonCTagPriority             int    `json:"usPonCTagPriority,omitempty"`
	UsPonSTagPriority             int    `json:"usPonSTagPriority,omitempty"`
	DsPonCTagPriority             int    `json:"dsPonCTagPriority,omitempty"`
	DsPonSTagPriority             int    `json:"dsPonSTagPriority,omitempty"`
	ServiceName                   string `json:"serviceName,omitempty"`
}

func (c *OnosClient) GetProgrammedSubscribers() ([]ProgrammedSubscriber, error) {
	method := http.MethodGet
	url := fmt.Sprintf("http://%s/onos/olt/oltapp/programmed-subscribers", c.endpoint)

	response, err := c.makeRequest(method, url)
	if err != nil {
		return nil, err
	}

	var subscribers struct {
		Entries []ProgrammedSubscriber `json:"entries"`
	}
	err = json.Unmarshal([]byte(response.Body), &subscribers)
	if err != nil {
		return nil, err
	}

	return subscribers.Entries, nil
}

///////////////////////////////////////////////////////////////////////// ONOS Core APIs

type OnosPort struct {
	Element     string            `json:"element"` //Device ID
	Port        string            `json:"port"`    //Port number
	IsEnabled   bool              `json:"isEnabled"`
	Type        string            `json:"type"`
	PortSpeed   uint              `json:"portSpeed"`
	Annotations map[string]string `json:"annotations"`
}

func (c *OnosClient) GetPorts() ([]OnosPort, error) {
	method := http.MethodGet
	url := fmt.Sprintf("http://%s/onos/v1/devices/ports", c.endpoint)

	response, err := c.makeRequest(method, url)
	if err != nil {
		return nil, err
	}

	var ports struct {
		Ports []OnosPort `json:"ports"`
	}
	err = json.Unmarshal([]byte(response.Body), &ports)
	if err != nil {
		return nil, err
	}

	return ports.Ports, nil
}

///////////////////////////////////////////////////////////////////////// ONOS SADIS APIs

type BandwidthProfile struct {
	Id  string `json:"id"`
	Cir int64  `json:"cir"`
	Cbs string `json:"cbs"`
	Air int64  `json:"air"`
	Gir int64  `json:"gir"`
	Eir int64  `json:"eir"`
	Ebs string `json:"ebs"`
	Pir int64  `json:"pir"`
	Pbs string `json:"pbs"`
}

func (c *OnosClient) GetBandwidthProfile(id string) (*BandwidthProfile, error) {
	method := http.MethodGet
	url := fmt.Sprintf("http://%s/onos/sadis/bandwidthprofile/%s", c.endpoint, id)

	response, err := c.makeRequest(method, url)
	if err != nil {
		return nil, err
	}

	var bwProfiles struct {
		Entry []BandwidthProfile `json:"entry"`
	}
	err = json.Unmarshal([]byte(response.Body), &bwProfiles)
	if err != nil {
		return nil, err
	}

	//The response has a list, but always returns one item
	//Verify this is correct and return it
	if len(bwProfiles.Entry) != 1 {
		return nil, fmt.Errorf("unexpected-number-of-bw-profile-entries: id=%s len=%d", id, len(bwProfiles.Entry))
	}

	return &bwProfiles.Entry[0], nil
}
