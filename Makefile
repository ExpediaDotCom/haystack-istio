.PHONY: deploy docker build

BINARY := haystackadapter

build:  server/main.go
	GOOS=linux GOARCH=amd64 go build -o ${BINARY} server/main.go
	chmod 755 ${BINARY}

docker:  build
	mv ${BINARY} docker/.
	docker build -f docker/Dockerfile -t haystack-istio-adapter:1 docker/

.PHONY: glide
glide:
	glide --version || go get github.com/Masterminds/glide
	glide update
	rm -rf vendor/istio.io/istio/vendor

.PHONY: validate
validate:
	./scripts/validate-go

deploy: docker
	kubectl -n istio-system apply -f haystack-adapter.yaml

