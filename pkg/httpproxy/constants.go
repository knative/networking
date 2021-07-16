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

package httpproxy

import "time"

// FlushInterval controls the time when we flush the connection in the
// reverse proxies (Activator, QP).
// NB: having it equal to 0 is a problem for streaming requests
// since the data won't be transferred in chunks less than 4kb, if the
// reverse proxy fails to detect streaming (gRPC, e.g.).
const FlushInterval = 20 * time.Millisecond
