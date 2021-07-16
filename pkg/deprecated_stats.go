/*
Copyright 2021 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pkg

import (
	"knative.dev/networking/pkg/stats"
)

type (
	// ReqEvent represents either an incoming or closed request.
	// Deprecated: use stats.ReqEvent
	ReqEvent = stats.ReqEvent

	// ReqEventType denotes the type (incoming/closed) of a ReqEvent.
	// Deprecated: use stats.ReqEventType
	ReqEventType = stats.ReqEventType

	// RequestStats collects statistics about requests as they flow in and out of the system.
	// Deprecated: use stats.RequestStats
	RequestStats = stats.RequestStats

	// RequestStatsReport are the metrics reported from the the request stats collector
	// at a given time.
	// Deprecated: use stats.RequestStatsReport
	RequestStatsReport = stats.RequestStatsReport
)

var (
	// NewRequestStats builds a RequestStats instance, started at the given time.
	// Deprecated: use stats.NewRequestStats
	NewRequestStats = stats.NewRequestStats
)

const (
	// ReqIn represents an incoming request
	// Deprecated: use stats.ReqIn
	ReqIn ReqEventType = stats.ReqIn

	// ReqOut represents a finished request
	// Deprecated: use stats.ReqOut
	ReqOut = stats.ReqOut

	// ProxiedIn represents an incoming request through a proxy.
	// Deprecated: use stats.ProxiedIn
	ProxiedIn = stats.ProxiedIn

	// ProxiedOut represents a finished proxied request.
	// Deprecated: use stats.ProxiedOut
	ProxiedOut = stats.ProxiedOut
)
