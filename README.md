# `terrarium`: Easy Python Containers

`terrarium` is a simple, fast container image builder for Python applications.

It's ideal for use cases where your project just needs to install some dependencies (e.g. via a `requirements.txt`).

`terrarium` builds images by effectively executing `pip install` on your local machine, and as such doesn't require `docker` to be installed, nor does it require `root` permissions.
This can make it a good fit for lightweight CI/CD use cases.

## Install `terrarium` and get started!

### Usage

```shell
# set the registry/repository that we want to push to
export TERRARIUM_DOCKER_REPO=registry.example.org/foo/bar
# set the OCI image that we will use as a base (this is the default)
export TERRARIUM_DEFAULT_BASE_IMAGE=python:3.12

# build the image
terrarium build . --tags v1.2.3 --entrypoint app.py

# the resulting image will be available at `registry.example.org/foo/bar:v1.2.3`
```

### Acknowledgements

This work is inspired by [`ko`](https://github.com/ko-build/ko) and [`nib`](https://github.com/djcass44/nib).
