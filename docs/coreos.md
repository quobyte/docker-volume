# Setup on CoreOS

Set the mount flags to `shared` this step shouldn't be needed with docker 1.12+.

```
$ sudo mkdir /etc/systemd/system/docker.service.d/
$ sudo sh -c 'echo -e "[Service]
MountFlags=shared
" > /etc/systemd/system/docker.service.d/slave-mount-flags.conf'

# Restart Docker
$ sudo systemctl daemon-reload
$ sudo systemctl restart docker
```

Create a directory to mount quobyte.

```
$ sudo mkdir /mnt/quobyte
```

Now mount Quobyte with Docker. If the command below fails use `-h $(hostname)` instead of `-h $(hostname -f)`.

```
$ docker run -d --name quobyte-client --privileged \
  -e QUOBYTE_REGISTRY=localhost:7861 \
  -p 55000:55000 \
  -v /mnt/quobyte:/quobyte:shared \
  -h $(hostname -f) \
  quay.io/quobyte/quobyte-client:1.4

# Validate mount
$ mount | grep quobyte
```

Create the config file for the plugin:

```
$ sudo mkdir /etc/quobyte
$ sudo cat /etc/quobyte/docker-quobyte.env
QUOBYTE_API_USER=admin
QUOBYTE_API_PASSWORD=quobyte
QUOBYTE_API_URL=http://quobyte:7860
QUOBYTE_REGISTRY=quobyte:7861
```

Create the systemd service

```
$ sudo cat /etc/systemd/system/docker-quobyte-plugin.service
[Unit]
Description=Docker Quobyte Plugin
Documentation=https://github.com/johscheuer/go-quobyte-docker
Before=docker.service
After=network.target docker.service
Requires=docker.service

[Service]
EnvironmentFile=/etc/quobyte/docker-quobyte.env
ExecStart=/opt/bin/docker-quobyte-plugin --user ${QUOBYTE_API_USER} --password ${QUOBYTE_API_PASSWORD} --api ${QUOBYTE_API_URL} --registry ${QUOBYTE_REGISTRY} --path /mnt/quobyte --group docker

[Install]
WantedBy=multi-user.target
```

Start the Plugin

```
$ sudo systemctl daemon-reload
$ sudo systemctl start docker-quobyte-plugin
$ sudo systemctl enable docker-quobyte-plugin
```
