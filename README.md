# Appvia Hub Kubernetes Agent

## Getting started
* Install [dep](https://github.com/golang/dep), download dependencies and build:
```
$ brew install dep
$ dep ensure
$ mkdir -p bin
$ go build -o bin/hub-kubernetes-agent
```
* Start the server
```
$ ./bin/hub-kubernetes-agent
```
