.PHONY: deploy docker build publish

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

deploy:
	kubectl -n istio-system apply -f haystack-adapter.yaml
