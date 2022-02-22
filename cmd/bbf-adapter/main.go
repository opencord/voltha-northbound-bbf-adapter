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
	"github.com/opencord/voltha-northbound-bbf-adapter/internal/config"
)

const (
	bbfAdapterService = "bbf-adapter-service"
)

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
	ctx := context.Background()
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

	probeCtx := context.WithValue(ctx, probe.ProbeContextKey, p)
	closer, err := log.GetGlobalLFM().InitTracingAndLogCorrelation(conf.TraceEnabled, conf.TraceAgentAddress, conf.LogCorrelationEnabled)
	if err != nil {
		logger.Warnw(ctx, "unable-to-initialize-tracing-and-log-correlation-module", log.Fields{"error": err})
	} else {
		defer log.TerminateTracing(closer)
	}

	//Configure readiness probe
	p.RegisterService(
		probeCtx,
		bbfAdapterService,
	)

	go func() {
		probe.UpdateStatusFromContext(probeCtx, bbfAdapterService, probe.ServiceStatusRunning)

		for {
			select {
			case <-ctx.Done():
				logger.Info(ctx, "Context closed")
				break
			case <-time.After(15 * time.Second):
				logger.Warn(ctx, "BBF Adapter currently has no implemented logic.")
			}
		}
	}()

	code := waitForExit(ctx)
	logger.Infow(ctx, "received-a-closing-signal", log.Fields{"code": code})

	elapsed := time.Since(start)
	logger.Infow(ctx, "run-time", log.Fields{"time": elapsed.Seconds()})
}
