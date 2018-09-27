#!/bin/bash

mkdir -p ${GOPATH}/src/istio.io/istio

git clone https://github.com/istio/istio ${GOPATH}/src/istio.io/istio && cd ${GOPATH}/src/istio.io/istio && git checkout bbee2cec0972aa221aa5464335aeeed8d87b5539

make -C ${GOPATH}/src/istio.io/istio mixs

if [[ "x${IS_TRAVIS}" == "xtrue" ]]; then
    mkdir -p ${GOPATH}/src/istio.io/istio/mixer/adapter/haystack
	cp -a ${GOPATH}/src/github.com/ExpediaDotCom/haystack-istio/* ${GOPATH}/src/istio.io/istio/mixer/adapter/haystack/
else 
    ln -s ${GOPATH}/src/github.com/ExpediaDotCom/haystack-istio ${GOPATH}/src/istio.io/istio/mixer/adapter/haystack
fi

echo "go get github.com/ExpediaDotCom/haystack-client-go..."
go get github.com/ExpediaDotCom/haystack-client-go

rm -rf ${GOPATH}/src/istio.io/istio/vendor/golang.org/x/net/trace #hack to avoid collisions in dependencies
