.PHONY: build-pip build-poetry

build-pip:
	TERRARIUM_DEFAULT_BASE_IMAGE=harbor.dcas.dev/docker.io/library/python:3.12 go run main.go build samples/sample-pip/ --entrypoint app.py --save /tmp/test.tar --v=10
	docker load < /tmp/test.tar

UID := $(shell id -u)
GID := $(shell id -u)

.ONESHELL:
build-poetry:
	mkdir -p bin/
	go build -o bin/terrarium main.go

	docker run \
		-e TERRARIUM_DEFAULT_BASE_IMAGE=harbor.dcas.dev/docker.io/library/python:3.12 \
		-v ./bin:/app \
		-v ./samples:/samples \
		--entrypoint /bin/bash harbor.dcas.dev/docker.io/library/python:3.12 \
		-c "/app/terrarium build /samples/sample-poetry --install-poetry --entrypoint app.py --save /app/test.tar --v=10"
	docker load < ./bin/test.tar