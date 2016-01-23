Quobyte volume plug-in for Docker
=================================

Setup:
* create a user in Quobyte for the plug-in:
```
 qmgmt -u <api-url> user config add docker <email>
```

* set mandatory configuration in quobyte.py
* sudo ./quobyte.py

Examples:

```
# docker create --driver quobyte --name <volumename> --opt volume_config=MyConfig
# docker create --driver quobyte --name <volumename>
# docker rm <volumename>
# docker run --volume-driver=quobyte -v <quobyte volumename>:path
```
