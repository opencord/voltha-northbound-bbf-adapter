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

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/opencord/voltha-lib-go/v7/pkg/log"
	"github.com/opencord/voltha-lib-go/v7/pkg/probe"
	"github.com/opencord/voltha-lib-go/v7/pkg/version"
	clients "github.com/opencord/voltha-northbound-bbf-adapter/internal/clients"
	"github.com/opencord/voltha-northbound-bbf-adapter/internal/config"
)

//String for readiness probe services
const (
	bbfAdapterService = "bbf-adapter-service"
)

type bbfAdapter struct {
	conf            *config.BBFAdapterConfig
	volthaNbiClient *clients.VolthaNbiClient
	oltAppClient    *clients.OltAppClient
}

func newBbfAdapter(conf *config.BBFAdapterConfig) *bbfAdapter {
	return &bbfAdapter{
		conf: conf,
	}
}

func (a *bbfAdapter) start(ctx context.Context, wg *sync.WaitGroup) {
	var err error

	//Connect to the voltha northbound api
	a.volthaNbiClient = clients.NewVolthaNbiClient(a.conf.VolthaNbiEndpoint)
	if err = a.volthaNbiClient.Connect(ctx, a.conf.TlsEnabled, a.conf.TlsVerify); err != nil {
		logger.Fatalw(ctx, "failed-to-open-voltha-nbi-grpc-connection", log.Fields{"err": err})
	} else {
		probe.UpdateStatusFromContext(ctx, a.conf.VolthaNbiEndpoint, probe.ServiceStatusRunning)
	}

	//Check if the REST APIs of the olt app are reachable
	a.oltAppClient = clients.NewOltAppClient(a.conf.OnosRestEndpoint, a.conf.OnosUser, a.conf.OnosPassword)
	if err := a.oltAppClient.CheckConnection(ctx); err != nil {
		logger.Fatalw(ctx, "failed-to-connect-to-onos-olt-app-api", log.Fields{"err": err})
	} else {
		probe.UpdateStatusFromContext(ctx, a.conf.OnosRestEndpoint, probe.ServiceStatusRunning)
	}

	//Run the main logic of the BBF adapter

	//Set the service as running, making the adapter finally ready
	probe.UpdateStatusFromContext(ctx, bbfAdapterService, probe.ServiceStatusRunning)
	logger.Info(ctx, "bbf-adapter-ready")

loop:
	for {
		select {
		case <-ctx.Done():
			logger.Info(ctx, "stop-for-context-done")
			break loop
		case <-time.After(15 * time.Second):
			//TODO: this is just to test functionality

			//Make a request to voltha
			devices, err := a.volthaNbiClient.Service.ListDevices(ctx, &empty.Empty{})
			if err != nil {
				logger.Errorw(ctx, "failed-to-list-devices", log.Fields{"err": err})
				continue
			}
			logger.Debugw(ctx, "Got devices from VOLTHA", log.Fields{"devNum": len(devices.Items)})

			//Make a request to Olt app
			response, err := a.oltAppClient.GetStatus()
			if err != nil {
				logger.Errorw(ctx, "failed-to-get-status", log.Fields{
					"err":     err,
					"reponse": response,
				})
				continue
			} else {
				logger.Debugw(ctx, "Got status from OltApp", log.Fields{"response": response})
			}

			logger.Warn(ctx, "BBF Adapter currently has no implemented logic.")
		}
	}

	probe.UpdateStatusFromContext(ctx, bbfAdapterService, probe.ServiceStatusStopped)
	wg.Done()
}

//Close all connections of the adapter
func (a *bbfAdapter) cleanup() {
	a.volthaNbiClient.Close()
}

func printBanner() {
	fmt.Println("  ____  ____  ______               _             _            ")
	fmt.Println(" |  _ \\|  _ \\|  ____|     /\\      | |           | |           ")
	fmt.Println(" | |_) | |_) | |__       /  \\   __| | __ _ _ __ | |_ ___ _ __ ")
	fmt.Println(" |  _ <|  _ <|  __|     / /\\ \\ / _` |/ _` | '_ \\| __/ _ \\ '__|")
	fmt.Println(" | |_) | |_) | |       / ____ \\ (_| | (_| | |_) | ||  __/ |   ")
	fmt.Println(" |____/|____/|_|      /_/    \\_\\__,_|\\__,_| .__/ \\__\\___|_|   ")
	fmt.Println("                                          | |                 ")
	fmt.Println("                                          |_|                 ")
}

func printVersion() {
	fmt.Println("VOLTHA Northbound BBF Adapter")
	fmt.Println(version.VersionInfo.String("  "))
}

func waitForExit(ctx context.Context) int {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	exitChannel := make(chan int)

	go func() {
		s := <-signalChannel
		switch s {
		case syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGQUIT:
			logger.Infow(ctx, "closing-signal-received", log.Fields{"signal": s})
			exitChannel <- 0
		default:
			logger.Infow(ctx, "unexpected-signal-received", log.Fields{"signal": s})
			exitChannel <- 1
		}
	}()

	code := <-exitChannel
	return code
}

func main() {
	ctx, cancelCtx := context.WithCancel(context.Background())

	start := time.Now()

	conf := config.LoadConfig(ctx)

	//Logging
	logLevel, err := log.StringToLogLevel(conf.LogLevel)
	if err != nil {
		logger.Fatalf(ctx, "Cannot setup logging, %s", err)
	}

	// Setup default logger - applies for packages that do not have specific logger set
	if _, err := log.SetDefaultLogger(log.JSON, logLevel, log.Fields{}); err != nil {
		logger.With(log.Fields{"error": err}).Fatal(ctx, "Cannot setup logging")
	}

	// Update all loggers (provisionned via init) with a common field
	if err := log.UpdateAllLoggers(log.Fields{}); err != nil {
		logger.With(log.Fields{"error": err}).Fatal(ctx, "Cannot setup logging")
	}

	log.SetAllLogLevel(logLevel)

	defer func() {
		err := log.CleanUp()
		if err != nil {
			logger.Errorw(context.Background(), "unable-to-flush-any-buffered-log-entries", log.Fields{"error": err})
		}
	}()

	// Print version and exit
	if conf.PrintVersion {
		printVersion()
		return
	}

	// Print banner if specified
	if conf.PrintBanner {
		printBanner()
	}

	logger.Infow(ctx, "config", log.Fields{"config": *conf})

	p := &probe.Probe{}
	go p.ListenAndServe(ctx, conf.ProbeAddress)

	//Register all services that will need to be initialized before considering the adapter ready
	probeCtx := context.WithValue(ctx, probe.ProbeContextKey, p)
	p.RegisterService(
		ctx,
		bbfAdapterService,
		conf.VolthaNbiEndpoint,
		conf.OnosRestEndpoint,
	)

	closer, err := log.GetGlobalLFM().InitTracingAndLogCorrelation(conf.TraceEnabled, conf.TraceAgentAddress, conf.LogCorrelationEnabled)
	if err != nil {
		logger.Warnw(ctx, "unable-to-initialize-tracing-and-log-correlation-module", log.Fields{"error": err})
	} else {
		defer log.TerminateTracing(closer)
	}

	adapter := newBbfAdapter(conf)

	//Run the adapter
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go adapter.start(probeCtx, wg)
	defer adapter.cleanup()

	//Wait a signal to stop execution
	code := waitForExit(ctx)
	logger.Infow(ctx, "received-a-closing-signal", log.Fields{"code": code})

	//Stop everything that waits for the context to be done
	cancelCtx()
	//Wait for the adapter logic to stop
	wg.Wait()

	elapsed := time.Since(start)
	logger.Infow(ctx, "run-time", log.Fields{"time": elapsed.Seconds()})
}
