[![Build Status](https://travis-ci.org/ExpediaDotCom/haystack-istio.svg?branch=master)](https://travis-ci.org/ExpediaDotCom/haystack-istio)
[![License](https://img.shields.io/badge/license-Apache%20License%202.0-blue.svg)](https://github.com/ExpediaDotCom/haystack/blob/master/LICENSE)

# Istio adpater for haystack distributed tracing

This is an Istio adapter that forwards the spans to [haystack](http://github.com/ExpediaDotCom/haystack) subsystem.
The istio's [mixer](https://github.com/istio/istio/tree/master/mixer) forwards the telemetry data to the adapter that runs as a grpc server. The adapter internally converts the istio's span object into protobuf object and forwards to the [haystack-agent](http://github.com/ExpediaDotCom/haystack-agent) running locally. The [k8s](./haystack-adapter.yaml) spec does this job of running the adapter and haystack-agent in the same k8s pod.

## How to deploy adapter?
Following steps are required to run the adapter:

1. kubectl apply -f haystack-adapter.yaml
2. kubectl apply -f [tracespan.yaml](https://github.com/istio/istio/blob/master/mixer/template/tracespan/tracespan.yaml) 
3. kubectl apply -f config/haystack.yaml -f operator/haystack-operator.yaml

The first step installs haystack-agent and adapter(grpc server). The second step installs the [tracespan](https://istio.io/docs/reference/config/policy-and-telemetry/templates/tracespan/) template in istio. You can skip this step if this template is already installed. The third step registers the haystack-adapter within istio, it configures the handler, instance object and rule. For more details read [this](https://istio.io/blog/2017/adapter-model/)  
 
## How to build this library?
`make glide` - if you are running for the very first time

`make build` - builds the adapter code

`make deploy` - will build the docker image and deploy in kubernetes cluster.




