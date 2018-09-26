.PHONY: deploy docker build

BINARY := haystackadapter

build:  server/main.go
	go build -o ${BINARY} server/main.go
	chmod 755 ${BINARY}

docker:  build
	mv ${BINARY} docker/.
	docker build -f docker/Dockerfile -t haystack-istio-adapter:1 docker/

.PHONY: setup
setup:
	mkdir -p ${GOPATH}/src/istio.io/istio
	git clone https://github.com/istio/istio ${GOPATH}/src/istio.io/istio && cd ${GOPATH}/src/istio.io/istio && git checkout bbee2cec0972aa221aa5464335aeeed8d87b5539
	make -C ${GOPATH}/src/istio.io/istio mixs
	ln -s ${GOPATH}/src/github.com/ExpediaDotCom/haystack-istio ${GOPATH}/src/istio.io/istio/mixer/adapter/haystack
	go get github.com/ExpediaDotCom/haystack-client-go
	rm -rf ${GOPATH}/src/istio.io/istio/vendor/golang.org/x/net/trace #hack to avoid collisions in dependencies

.PHONY: validate
validate:
	./scripts/validate-go

deploy: docker
	kubectl -n istio-system apply -f haystack-adapter.yaml
