[![Build Status](https://travis-ci.org/quobyte/docker-volume.svg?branch=master)](https://travis-ci.org/quobyte/docker-volume)

# Quobyte volume plug-in for Docker

This plugin allows you to use [Quobyte](https://www.quobyte.com) with Docker without installing the Quobyte client on the host system (e.q. Rancher/CoreOS) for more information look at [docs/coreos.md](docs/coreos.md).

## Tested

OS              | Docker Version
--------------- | :------------:
CentOS 7.2      |     1.10.3
Ubuntu 16.04    |     1.11.2
Ubuntu 16.04    |     1.12.0
CoreOS 1097.0.0 |     1.11.2

## Setup

### Binary

Download the binary and the systemd files from the [releases](https://github.com/quobyte/docker-volume/releases) page. To download version `1.1` of this plugin run the following step:

```bash
curl -LO https://github.com/quobyte/docker-volume/releases/download/v1.1/docker-quobyte-plugin.tar.gz
tar xfz docker-quobyte-plugin.tar.gz
```

### Create a user in Quobyte for the plug-in:

This step is optional and should only be used if Quobytes build-in user database is used.

```
$ qmgmt -u <api-url> user config add docker <email>
```

### Systemd Files

Install systemd files and set your variables in systemd/docker-quobyte.env.sample

```
$ cp systemd/docker-quobyte.env.sample /etc/quobyte/docker-quobyte.env
$ cp bin/docker-quobyte-plugin /usr/local/bin/
$ cp systemd/docker-quobyte-plugin.service /lib/systemd/system

$ systemctl daemon-reload
$ systemctl start docker-quobyte-plugin
$ systemctl enable docker-quobyte-plugin
$ systemctl status docker-quobyte-plugin
```

### Usage

```
$ bin/docker-quobyte-plugin -h
Usage of bin/docker-quobyte-plugin:
  -api string
      URL to the API server(s) in the form http(s)://host[:port][,host:port] or SRV record name (default "http://localhost:7860")
  -configuration_name string
      Name of the volume configuration of new volumes (default "BASE")
  -group string
      Group to create the unix socket (default "root")
  -max-fs-checks int
      Maximimum number of filesystem checks when a Volume is created before returning an error (default 5)
  -max-wait-time float
      Maximimum wait time for filesystem checks to complete when a Volume is created before returning an error (default 30)
  -options string
      Fuse options to be used when Quobyte is mounted (default "-o user_xattr")
  -password string
      Password for the user to connect to the Quobyte API server (default "quobyte")
  -path string
      Path where Quobyte is mounted on the host (default "/run/docker/quobyte/mnt")
  -registry string
      URL to the registry server(s) in the form of host[:port][,host:port] or SRV record name (default "localhost:7861")
  -tenant_id string
      Id of the Quobyte tenant in whose domain the operation takes place (default "no default")
  -user string
      User to connect to the Quobyte API server (default "root")
  -version
      Shows version string
```

## Examples

### Create a volume

```
$ docker volume create --driver quobyte --name <volumename> --opt tenant_id=<your Quobyte tenant_id>
# Set user and group of the volume
$ docker volume create --driver quobyte --name <volumename> --opt user=docker --opt group=docker --opt tenant_id=<your Quobyte tenant_id>
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

## Development

### Build

Get the code

```
$ go get -u github.com/quobyte/docker-volume
```

#### Dependency Management

For the dependency management we use [golang dep](https://github.com/golang/dep)

#### Linux

```
$ go build -ldflags "-s -w" -o bin/docker-quobyte-plugin .
```

#### OSX/MacOS

```
$ GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o bin/docker-quobyte-plugin .
```

#### Docker

```
$ docker run --rm -v "$GOPATH":/work -e "GOPATH=/work" -w /work/src/github.com/quobyte/docker-volume golang:1.8 go build -v -ldflags "-s -w" -o bin/quobyte-docker-plugin
```
