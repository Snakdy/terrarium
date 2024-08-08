#!/usr/bin/env bash

set -eux

export TERRARIUM_DEFAULT_BASE_IMAGE=harbor.dcas.dev/registry.gitlab.com/av1o/base-images/python-3.12:latest

# build the binary
mkdir -p bin/
go build -o bin/terrarium main.go

# setup the test function
function build_pyapp_docker() {
	docker run \
		-e TERRARIUM_DEFAULT_BASE_IMAGE="$TERRARIUM_DEFAULT_BASE_IMAGE" \
		-v ./bin:/app \
		--entrypoint /bin/bash harbor.dcas.dev/registry.gitlab.com/av1o/base-images/python-3.12:latest \
		-c "git clone $1 /tmp/workspace && /app/terrarium build /tmp/workspace/ --install-poetry --entrypoint $2 --save /app/test.tar --v=10"
	docker load < ./bin/test.tar
}

function clone() {
	rm -rf "tests/$2"
	git clone "$1" "tests/$2"
}

# reset if we need to and
# clone the test repositories


# run the tests
build_pyapp_docker https://gitlab.dcas.dev/autodevops/python-sample.git app.py
docker run -it image