# docker-machine-pwd-priver

Docker machine PWD driver


## Getting started

This driver tricks machine and allows to create / remove PWD instances remotely. 
As PWD uses static certificates, it's mandatory to set a different machine storage path
to avoid the driver overwrite your current docker-machine certs


Before using it please make sure of the following:

- Set $MACHINE_STORAGE_PATH env variable to an *existing* directory (i.e /tmp/pwd)
- Create a session in PWD and set PWD_SESSION_ID env variable or use --pwd-session-id flag when creating an instance


## Installing

### Easy way

Download the binary for your platform from the [releases]("https://github.com/franela/play-with-docker/releases") section and place it somewhere in your PATH


### Hard way

Use `go get github.com/franela/docker-machine-driver-pwd` and make sure that
`docker-machine-driver-pwd` is located somwhere in your PATH



## Usage

Creating an instance:

```
# Create a session in play-with-docker.com and set env variable
export MACHINE_STORAGE_PATH="/tmp/pwd"
export PWD_SESSION_ID="my-session-id"
docker-machine create -d pwd node1
eval $(docker-machine node1)
docker ps
```


Remove an instance


```
# Make sure PWD_SESSION_ID is still set
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
