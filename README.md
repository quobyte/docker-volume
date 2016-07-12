# Quobyte volume plug-in for Docker

Tested with `CentOS 7.2` and `Docker 1.10.3`. This plugin allows you to use [Quobyte](www.quobyte.com) with Docker without installing the Quobyte client on the host system (e.q. Rancher/CoreOS).

## Build

Get the code

```
$ go get -u github.com/quobyte/api
$ go get -u github.com/quobyte/docker-volume
```

### Linux

```
$ go build -o docker-quobyte-plugin .
$ cp quobyte-docker-plugin /usr/libexec/docker/docker-quobyte-plugin
```

### OSX/MacOS

```
$ GOOS=linux GOARCH=amd64 go build -o docker-quobyte-plugin
$ cp quobyte-docker-plugin /usr/libexec/docker/docker-quobyte-plugin
```

## Setup

### Create a user in Quobyte for the plug-in:

This step is optional.

```
$ qmgmt -u <api-url> user config add docker <email>
```

### Set mandatory configuration in environment

```
$ export QUOBYTE_API_USER=docker
$ export QUOBYTE_API_PASSWORD=...
$ export QUOBYTE_API_URL=http://<host>:7860/
# host[:port][,host:port] or SRV record name
$ export QUOBYTE_REGISTRY=quobyte.corp
```

### Install systemd files Set the variables in systemd/docker-quobyte.env.sample

```
$ cp systemd/docker-quobyte.env.sample /etc/quobyte/docker-quobyte.env
$ cp docker-quobyte-plugin /usr/libexec/docker/
$ cp systemd/* /lib/systemd/system

$ systemctl daemon-reload
$ systemctl start docker-quobyte-plugin
$ systemctl enable docker-quobyte-plugin
$ systemctl status docker-quobyte-plugin
```

## Examples

### Create a volume

```
$ docker volume create --driver quobyte --name <volumename>
# Set user and group of the volume
$ docker volume create --driver quobyte --name <volumename> --opt user=docker --opt group=docker
```

### Delete a volume

```
$ docker volume rm <volumename>
```

### List all volumes

```
$ docker volume ls
```

### Attach volume to container

```
$ docker run --volume-driver=quobyte -v <volumename>:/vol busybox sh -c 'echo "Hello World" > /vol/hello.txt'
```
