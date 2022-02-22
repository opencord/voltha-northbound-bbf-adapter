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

package config

import (
	"context"
	"flag"
)

type BBFAdapterConfig struct {
	PrintVersion          bool
	PrintBanner           bool
	LogLevel              string
	ProbeAddress          string
	TraceEnabled          bool
	TraceAgentAddress     string
	LogCorrelationEnabled bool
}

// LoadConfig loads the BBF adapter configuration through
// default values and CLI arguments
func LoadConfig(ctx context.Context) *BBFAdapterConfig {
	conf := getDefaultConfig()

	flag.StringVar(&conf.LogLevel, "log_level", conf.LogLevel, "Log level (DEBUG, INFO, WARN, ERROR)")
	flag.BoolVar(&conf.PrintVersion, "version", conf.PrintVersion, "Print the version and exit")
	flag.BoolVar(&conf.PrintBanner, "banner", conf.PrintBanner, "Print the banner at startup")
	flag.StringVar(&conf.ProbeAddress, "probe_address", conf.ProbeAddress, "The address on which to listen to answer liveness and readiness probe queries over HTTP")
	flag.BoolVar(&conf.TraceEnabled, "trace_enabled", conf.TraceEnabled, "Whether to send logs to tracing agent")
	flag.StringVar(&conf.TraceAgentAddress, "trace_agent_address", conf.TraceAgentAddress, "The address of tracing agent to which span info should be sent")
	flag.BoolVar(&conf.LogCorrelationEnabled, "log_correlation_enabled", conf.LogCorrelationEnabled, "Whether to enrich log statements with fields denoting operation being executed for achieving correlation")

	flag.Parse()

	return conf
}

// getDefaultConfig returns a BBF Adapter configuration with default values
func getDefaultConfig() *BBFAdapterConfig {
	return &BBFAdapterConfig{
		LogLevel:              "ERROR",
		PrintVersion:          false,
		PrintBanner:           false,
		ProbeAddress:          ":8080",
		TraceEnabled:          false,
		TraceAgentAddress:     "127.0.0.1:6831",
		LogCorrelationEnabled: true,
	}
}
