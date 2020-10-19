/*
Copyright 2020 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ingress

import (
	"knative.dev/networking/test"
)

// RunConformance will run ingress conformance tests
//
// Depending on the options it may test alpha and beta features
func RunConformance(t *test.T) {
	t.Stable("basics", TestBasics)
	t.Stable("basics/http2", TestBasicsHTTP2)
	t.Stable("grpc", TestGRPC)
	t.Stable("grpc/split", TestGRPCSplit)
	t.Stable("headers/pre-split", TestPreSplitSetHeaders)
	t.Stable("headers/post-split", TestPostSplitSetHeaders)
	t.Stable("hosts/multiple", TestMultipleHosts)
	t.Stable("dispatch/path", TestPath)
	t.Stable("dispatch/percentage", TestPercentage)
	t.Stable("dispatch/path_and_percentage", TestPathAndPercentageSplit)
	t.Stable("timeout", TestTimeout)
	t.Stable("tls", TestIngressTLS)
	t.Stable("update", TestUpdate)
	t.Stable("visibility", TestVisibility)
	t.Stable("visibility/split", TestVisibilitySplit)
	t.Stable("visibility/path", TestVisibilityPath)
	t.Stable("ingressclass", TestIngressClass)
	t.Stable("websocket", TestWebsocket)
	t.Stable("websocket/split", TestWebsocketSplit)

	t.Beta("headers/probe", TestProbeHeaders)

	t.Alpha("headers/tag", TestTagHeaders)
	t.Alpha("host-rewrite", TestRewriteHost)
}
