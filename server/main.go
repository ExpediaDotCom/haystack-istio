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
