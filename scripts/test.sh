#!/usr/bin/env bash

set -eux

export TERRARIUM_DEFAULT_BASE_IMAGE=harbor.dcas.dev/docker.io/library/python:3.12

# build the binary
mkdir -p bin/
go build -o bin/terrarium main.go

# setup the test function
function build_pyapp_docker() {
	docker run \
		-e TERRARIUM_DEFAULT_BASE_IMAGE="$TERRARIUM_DEFAULT_BASE_IMAGE" \
		-v ./bin:/app \
		-v ./tests:/workspace \
		--entrypoint /bin/bash harbor.dcas.dev/docker.io/library/python:3.12 \
		-c "pip install poetry && /app/terrarium build /workspace/$1 --entrypoint $2 --save /app/test.tar --v=10"
	docker load <./bin/test.tar
}

function clone() {
	rm -rf "tests/$2"
	git clone "$1" "tests/$2"
}

# reset if we need to and
# clone the test repositories
clone https://github.com/guenterfischer/python-app-template simple-poetry
clone https://github.com/Shopify/sample-django-app django
clone https://github.com/miguelgrinberg/microblog flask

# run the tests
build_pyapp_docker simple-poetry pyapp/main.py
docker run -it image --help

build_pyapp_docker django sample_django_app/manage.py
docker run -it image

build_pyapp_docker flask microblog.py
docker run image
