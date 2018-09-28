/*
 *  Copyright 2018 Expedia, Inc.
 *
 *     Licensed under the Apache License, Version 2.0 (the "License");
 *     you may not use this file except in compliance with the License.
 *     You may obtain a copy of the License at
 *
 *         http://www.apache.org/licenses/LICENSE-2.0
 *
 *     Unless required by applicable law or agreed to in writing, software
 *     distributed under the License is distributed on an "AS IS" BASIS,
 *     WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *     See the License for the specific language governing permissions and
 *     limitations under the License.
 *
 */

package main

import (
	"flag"
	"fmt"
	"os"

	a "istio.io/istio/mixer/adapter/haystack/adapter"
)

func main() {
	var port = flag.String("port", "0", "port for adapter grpc server")
	var agentHost = flag.String("agentHost", "localhost", "hostname where haystack-agent is running")
	var agentPort = flag.Int("agentPort", 35000, "port on which haystack-agent is listening")

	flag.Parse()

	s, err := a.NewHastackGrpcAdapter(*port, *agentHost, *agentPort)
	if err != nil {
		fmt.Printf("unable to start server: %v", err)
		os.Exit(-1)
	}

	shutdown := make(chan error, 1)
	go func() {
		s.Run(shutdown)
	}()
	_ = <-shutdown
}
