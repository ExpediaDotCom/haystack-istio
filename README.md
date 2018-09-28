[![Build Status](https://travis-ci.org/ExpediaDotCom/haystack-istio.svg?branch=master)](https://travis-ci.org/ExpediaDotCom/haystack-istio)
[![License](https://img.shields.io/badge/license-Apache%20License%202.0-blue.svg)](https://github.com/ExpediaDotCom/haystack/blob/master/LICENSE)

# Istio adapter for haystack distributed tracing

This is an Istio adapter that forwards the telemetry data to [haystack](http://github.com/ExpediaDotCom/haystack) system.

## How it works?
Istio's [mixer](https://istio.io/docs/concepts/policies-and-telemetry/) receives the telemtry data from envoy proxy that runs as a sidecar with microservice app. Mixer can be configured to forward this data to various adpaters. We have built a new adapter for haystack that runs as an out-of-process grpc server and can receive telemetry data from mixer. 

The adapter internally converts the istio's span object into [protobuf](https://github.com/ExpediaDotCom/haystack-idl/blob/master/proto/span.proto) format and forwards to [haystack-agent](http://github.com/ExpediaDotCom/haystack-agent). The haystack-agent runs as a sidecar with the adapter. You can run as many replicas in order to scale. 

We also provide [k8s](./haystack-adapter.yaml) spec that helps you deploy haystack-adapter and haystack-agent in the same pod.

## How to deploy adapter?
Following steps are required to run the adapter:

1. kubectl apply -f haystack-adapter.yaml
2. kubectl apply -f [tracespan.yaml](https://github.com/istio/istio/blob/master/mixer/template/tracespan/tracespan.yaml) 
3. kubectl apply -f config/haystack.yaml -f operator/haystack-operator.yaml

The first step installs haystack-agent and adapter(grpc server). The second step installs the [tracespan](https://istio.io/docs/reference/config/policy-and-telemetry/templates/tracespan/) template in istio. You can skip this step if this template is already installed. The third step registers the haystack-adapter within istio, it configures the handler, instance object and rule. For more details read [this](https://istio.io/blog/2017/adapter-model/)  
 
## How to build this library?
`make setup` - if you are running for the very first time. This does some hacking to setup the right environment

`make -C $GOPATH/src/istio.io/istio/mixer/adapter/haystack build` - builds the adapter code

`make -C $GOPATH/src/istio.io/istio/mixer/adapter/haystack docker deploy` - will build the docker image and deploy in kubernetes cluster.

## How to run integration tests?
Install docker and docker-compose. Add following entries in /etc/hosts
$(docker-machine ip) mixs
$(docker-machine ip) kafkasvc

`make -C $GOPATH/src/istio.io/istio/mixer/adapter/haystack integration_tests`

