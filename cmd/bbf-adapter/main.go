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
	"syscall"
	"time"

	"github.com/opencord/voltha-lib-go/v7/pkg/log"
	"github.com/opencord/voltha-lib-go/v7/pkg/probe"
	"github.com/opencord/voltha-lib-go/v7/pkg/version"
	"github.com/opencord/voltha-northbound-bbf-adapter/internal/clients"
	"github.com/opencord/voltha-northbound-bbf-adapter/internal/config"
	"github.com/opencord/voltha-northbound-bbf-adapter/internal/core"
	"github.com/opencord/voltha-northbound-bbf-adapter/internal/sysrepo"
)

//String for readiness probe services
const (
	bbfAdapterService = "bbf-adapter-service"
	sysrepoService    = "sysrepo"
)

type bbfAdapter struct {
	conf            *config.BBFAdapterConfig
	volthaNbiClient *clients.VolthaNbiClient
	oltAppClient    *clients.OltAppClient
	sysrepoPlugin   *sysrepo.SysrepoPlugin
}

func newBbfAdapter(conf *config.BBFAdapterConfig) *bbfAdapter {
	return &bbfAdapter{
		conf: conf,
	}
}

func (a *bbfAdapter) start(ctx context.Context) {
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

	//Create the global adapter that will be used by callbacks
	core.AdapterInstance = core.NewVolthaYangAdapter(a.volthaNbiClient, a.oltAppClient)

	//Load sysrepo plugin
	a.sysrepoPlugin, err = sysrepo.StartNewPlugin(ctx, a.conf.SchemaMountFilePath)
	if err != nil {
		logger.Fatalw(ctx, "failed-to-start-sysrepo-plugin", log.Fields{"err": err})
	} else {
		probe.UpdateStatusFromContext(ctx, sysrepoService, probe.ServiceStatusRunning)
	}

	//Set the service as running, making the adapter finally ready
	probe.UpdateStatusFromContext(ctx, bbfAdapterService, probe.ServiceStatusRunning)
	logger.Info(ctx, "bbf-adapter-ready")
}

//Close all connections of the adapter
func (a *bbfAdapter) cleanup(ctx context.Context) {
	core.AdapterInstance = nil

	a.volthaNbiClient.Close(ctx)

	err := a.sysrepoPlugin.Stop(ctx)
	if err != nil {
		logger.Errorw(ctx, "failed-to-stop-sysrepo-plugin", log.Fields{"err": err})
	}

	probe.UpdateStatusFromContext(ctx, bbfAdapterService, probe.ServiceStatusStopped)
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
		sysrepoService,
	)

	closer, err := log.GetGlobalLFM().InitTracingAndLogCorrelation(conf.TraceEnabled, conf.TraceAgentAddress, conf.LogCorrelationEnabled)
	if err != nil {
		logger.Warnw(ctx, "unable-to-initialize-tracing-and-log-correlation-module", log.Fields{"error": err})
	} else {
		defer log.TerminateTracing(closer)
	}

	adapter := newBbfAdapter(conf)

	//Run the adapter
	adapter.start(probeCtx)
	defer adapter.cleanup(probeCtx)

	//Wait a signal to stop execution
	code := waitForExit(ctx)
	logger.Infow(ctx, "received-a-closing-signal", log.Fields{"code": code})

	//Stop everything that waits for the context to be done
	cancelCtx()

	elapsed := time.Since(start)
	logger.Infow(ctx, "run-time", log.Fields{"time": elapsed.Seconds()})
}
