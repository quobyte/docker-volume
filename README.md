Quobyte volume plug-in for Docker
=================================

Setup:
* create a user in Quobyte for the plug-in:
```
 qmgmt -u <api-url> user config add docker <email>
```

* set mandatory configuration in quobyte-docker-volume.py
```
QMGMT_USER = "docker"
QMGMT_PASSWORD = ""
QUOBYTE_API_URL = "http://<host>:7860/"
# host[:port][,host:port] or SRV record name
QUOBYTE_REGISTRY = ""
```

* Start the plug-in with
``` 
sudo ./quobyte-docker-volume.py 
```

Examples:

```
# docker create --driver quobyte --name <volumename> --opt volume_config=MyConfig
# docker create --driver quobyte --name <volumename>
# docker rm <volumename>
# docker run --volume-driver=quobyte -v <quobyte volumename>:path
```

Todos:
* Add quobyte-docker-volume.service for systemd
