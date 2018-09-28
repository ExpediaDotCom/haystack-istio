.PHONY: deploy docker build publish integration_tests

BINARY := haystackadapter

build:  server/main.go
	GOOS=linux GOARCH=amd64 go build -o ${BINARY} server/main.go
	chmod 755 ${BINARY}

docker:
	mv ${BINARY} docker/.
	docker build -f docker/Dockerfile -t haystack-istio-adapter docker/

.PHONY: setup
setup:
	./scripts/setup.sh

.PHONY: validate
validate:
	./scripts/validate-go

publish: docker
	./scripts/publish-to-docker-hub.sh

integration_tests: build docker
	docker-compose -f docker/docker-compose.yaml -p sandbox up -d
	sleep 45
	go test ./...
	docker-compose -f docker/docker-compose.yaml -p sandbox stop
	docker rm -f $(shell docker ps -a -q)

deploy:
	kubectl -n istio-system apply -f haystack-adapter.yaml
	kubectl -n istio-system apply -f testdata/tracespan.yaml
	kubectl -n istio-system apply -f config/haystack.yaml
	kubectl -n istio-system apply -f testdata/haystack-operator.yaml
