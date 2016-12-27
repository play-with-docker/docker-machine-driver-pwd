# docker-machine-pwd-priver

Docker machine PWD driver


## Getting started

This driver tricks machine and allows to create / remove PWD instances remotely. 

Before using it please make sure of the following:

- Create a session in PWD and set PWD_SESSION_ID env variable or use --pwd-session-id flag when creating an instance


## Installing

### Easy way

Download the release bundle from the [releases](https://github.com/franela/play-with-docker/releases) section and place the binary that corresponds to your platform it somewhere in your PATH



### Hard way

Use `go get github.com/franela/docker-machine-driver-pwd` and make sure that
`docker-machine-driver-pwd` is located somwhere in your PATH



## Usage

Creating an instance:

```
# Create a session in play-with-docker.com and set env variable
docker-machine create -d pwd --pwd-session-id <your pwd session id> node1
eval $(docker-machine node1)
docker ps
```

Alternatively you can set the env variable `PWD_SESSION_ID` to avoid passing it as a flag every time.


Remove an instance


```
docker-machine rm -f node1
```

## Development

For local development it's necessary to set `PWD_PORT`, `PWD_HOSTNAME` and `PWD_SSL_PORT`
accordingly to use with local PWD.

i.e:

```
export PWD_PORT=3000
export PWD_SSL_PORT=3001
export PWD_HOSTNAME=localhost
```
