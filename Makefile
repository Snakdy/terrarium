.PHONY: build

build:
	TERRARIUM_DEFAULT_BASE_IMAGE=harbor.dcas.dev/docker.io/library/python:3.12 go run main.go build sample/ --entrypoint app.py --save /tmp/test.tar --v=10
	docker load < /tmp/test.tar