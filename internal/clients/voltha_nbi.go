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
	"crypto/tls"
	"fmt"
	"time"

	vgrpc "github.com/opencord/voltha-lib-go/v7/pkg/grpc"
	"github.com/opencord/voltha-lib-go/v7/pkg/log"
	"github.com/opencord/voltha-protos/v5/go/voltha"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
)

const (
	nbiInitialBackoffInterval = time.Second
	nbiMaxBackoffInterval     = time.Second * 10
)

//Used to keep track of a connection to a grpc endpoint of the northbound api
type VolthaNbiClient struct {
	conn     *grpc.ClientConn
	Service  voltha.VolthaServiceClient
	endpoint string
}

// Creates a new voltha northbound client
func NewVolthaNbiClient(endpoint string) *VolthaNbiClient {
	return &VolthaNbiClient{
		endpoint: endpoint,
	}
}

// Dials the grpc connection to the endpoint and sets the service as running
func (c *VolthaNbiClient) Connect(ctx context.Context, useTls bool, verifyTls bool) error {
	var opts []grpc.DialOption

	backoffConfig := backoff.DefaultConfig
	backoffConfig.MaxDelay = nbiMaxBackoffInterval

	opts = append(opts,
		grpc.WithConnectParams(
			grpc.ConnectParams{
				Backoff: backoffConfig,
			},
		),
	)

	if useTls {
		//TODO: should this be expanded with the ability to provide certificates?
		creds := credentials.NewTLS(&tls.Config{InsecureSkipVerify: !verifyTls})
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	logger.Debugw(ctx, "connecting-to-voltha-nbi-grpc", log.Fields{"endpoint": c.endpoint})

	var err error
	c.conn, err = grpc.DialContext(ctx, c.endpoint, opts...)

	if err != nil {
		return err
	}

	//Wait for the connection to be successful, with periodic updates on its status
	backoff := vgrpc.NewBackoff(nbiInitialBackoffInterval, nbiMaxBackoffInterval, vgrpc.DefaultBackoffMaxElapsedTime)
	for {
		if state := c.conn.GetState(); state == connectivity.Ready {
			break
		} else {
			logger.Warnw(ctx, "voltha-nbi-grpc-not-ready", log.Fields{"state": state})
		}

		if err := backoff.Backoff(ctx); err != nil {
			return fmt.Errorf("voltha-nbi-connection-stopped-due-to-context-done")
		}
	}

	c.Service = voltha.NewVolthaServiceClient(c.conn)

	logger.Debug(ctx, "voltha-nbi-grpc-connected")

	return nil
}

// Closes the connection and cleans up
func (c *VolthaNbiClient) Close(ctx context.Context) {
	c.conn.Close()
	c.Service = nil

	logger.Debug(ctx, "closed-voltha-nbi-grpc-connection")
}
