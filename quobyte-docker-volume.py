#!/usr/bin/python2.7
#
# A Docker volume plug-in
#
# Copyright 2016 Quobyte Inc. All rights reserved.
#
# Usage:
#   - set mandatory configuration (see below)
#   - Run as root
#
# Examples:
#   docker create --driver quobyte --name <volumename> --opt volume_config=MyConfig
#   docker create --driver quobyte --name <volumename>
#   docker rm <volumename>
#   docker run --volume-driver=quobyte -v <quobyte volumename>:path

from BaseHTTPServer import BaseHTTPRequestHandler
import BaseHTTPServer
import json
import os
import os.path
import socket
import time


def getenv_mandatory(name):
    result = os.getenv(name)
    if not result:
        raise BaseException("Please set " + name + " in environment")
    return result

# Mandatory configuration
QMGMT_USER = getenv_mandatory("QUOBYTE_API_USER")
QMGMT_PASSWORD = getenv_mandatory("QUOBYTE_API_PASSWORD")
QUOBYTE_API_URL = getenv_mandatory("QUOBYTE_API_URL")
# host[:port][,host:port] or SRV record name
QUOBYTE_REGISTRY = getenv_mandatory("QUOBYTE_REGISTRY")

# Optional configuration
MOUNT_QUOBYTE_PATH = ""
MOUNT_QUOBYTE_OPTIONS = "-o user_xattr"
QMGMT_PATH = ""
DEFAULT_VOLUME_CONFIGURATION = "BASE"


# Constants
PLUGIN_DIRECTORY = '/run/docker/plugins/'
PLUGIN_SOCKET = PLUGIN_DIRECTORY + 'quobyte.sock'
MOUNT_DIRECTORY = '/run/docker/quobyte/mnt'


def read_optional_config():
    global MOUNT_QUOBYTE_PATH
    MOUNT_QUOBYTE_PATH = os.getenv("MOUNT_QUOBYTE_PATH")
    global QMGMT_PATH
    QMGMT_PATH = os.getenv("QMGMT_PATH")
    options = os.getenv("MOUNT_QUOBYTE_OPTIONS")
    if options:
        global MOUNT_QUOBYTE_OPTIONS
        MOUNT_QUOBYTE_OPTIONS = options
    config = os.getenv("DEFAULT_VOLUME_CONFIGURATION")
    if config:
        global DEFAULT_VOLUME_CONFIGURATION
        DEFAULT_VOLUME_CONFIGURATION = config


def mount_all():
    binary = "mount.quobyte"
    if MOUNT_QUOBYTE_PATH:
        binary = os.path.join(MOUNT_QUOBYTE_PATH, binary)
    mnt_cmd = (binary + " " + MOUNT_QUOBYTE_OPTIONS + " " +
               QUOBYTE_REGISTRY + "/ " + MOUNT_DIRECTORY)
    print mnt_cmd
    return os.system(mnt_cmd)


def qmgmt(params):
    binary = "qmgmt"
    if not QUOBYTE_API_URL:
        print "Please configure API URL"
        raise Exception()
    if QMGMT_PATH:
        binary = os.path.join(QMGMT_PATH, binary)
    cmdline = binary + " -u " + QUOBYTE_API_URL + " " + params
    print cmdline
    exitcode = os.system(cmdline)
    print "==", exitcode
    return exitcode == 0


def volume_create(name, volume_config):
    return qmgmt(
        "volume create " +
        name +
        " root root " +
        volume_config +
        " 777")


def volume_delete(name):
    return qmgmt("volume delete -f " + name)


def volume_exists(name):
    return qmgmt("volume resolve " + name)


def is_mounted(mountpath):
    mounts = open('/proc/mounts')
    for mount in mounts:
        if mount.split()[1] == mountpath:
            return True
    return False


class UDSServer(BaseHTTPServer.HTTPServer):
    address_family = socket.AF_UNIX
    socket_type = socket.SOCK_STREAM

    def __init__(self, server_address, RequestHandlerClass):
        try:
            os.unlink(server_address)
        except OSError:
            if os.path.exists(server_address):
                raise
        self.socket = socket.socket(self.address_family, self.socket_type)
        BaseHTTPServer.HTTPServer.__init__(
            self, server_address, RequestHandlerClass)

    def server_bind(self):
        self.socket.bind(self.server_address)

    def server_activate(self):
        self.socket.listen(1)

    def server_close(self):
        self.socket.close()

    def fileno(self):
        return self.socket.fileno()

    def close_request(self, request):
        request.close()

    def get_request(self):
        return self.socket.accept()[0], 'uds'


class DockerHandler(BaseHTTPRequestHandler):
    mount_paths = {}

    def get_request(self):
        length = int(self.headers['content-length'])
        return json.loads(self.rfile.read(length))

    def respond(self, msg):
        self.send_response(200)
        self.send_header(
            "Content-type",
            "application/vnd.docker.plugins.v1+json")
        self.end_headers()
        print "Responding with", json.dumps(msg)
        self.wfile.write(json.dumps(msg))

    def do_POST(self):
        print self.get_request()
        if self.path == "/Plugin.Activate":
            self.respond({"Implements": ["VolumeDriver"]})
        elif self.path == "/VolumeDriver.Create":
            self.volume_driver_create()
        elif self.path == "/VolumeDriver.Remove":
            self.volume_driver_mount()
        elif self.path == "/VolumeDriver.Path" or self.path == "/VolumeDriver.Mount":
            self.volume_driver_mount()
        elif self.path == "/VolumeDriver.Get":
            self.volume_driver_list()
        elif self.path == "/VolumeDriver.Unmount":
            self.respond({"Err": ""})
        elif self.path == "/VolumeDriver.List":
            self.volume_driver_list()
        else:
            print "Unknown API operation:", self.path
            self.respond({"Err": "Unknown API operation: " + self.path})

    def volume_driver_create(self):
        volume_config = DEFAULT_VOLUME_CONFIGURATION
        request = self.get_request()
        name = request["Name"]
        if 'Opts' in request and request[
                'Opts'] and 'volume_configuration' in request['Opts']:
            volume_config = request['Opts']['volume_configuration']

        volume_create(name, volume_config)
        mountpoint = os.path.join(MOUNT_DIRECTORY, name)
        while not os.path.exists(mountpoint):
            print "Waiting for", mountpoint
            time.sleep(1)
        self.respond({"Err": ""})

    def volume_driver_remove(self):
        name = self.get_request()["Name"]
        if not volume_exists(name):
            self.respond({"Err": ""})
            return
        if volume_delete(name):
            self.respond({"Err": ""})
        else:
            self.respond({"Err": "Could not delete " + name})

    def volume_driver_mount(self):
        name = self.get_request()["Name"]
        mountpoint = os.path.join(MOUNT_DIRECTORY, name)
        if os.path.exists(mountpoint):
            self.respond({"Err": "", "Mountpoint": mountpoint})
        else:
            self.respond({"Err": "Not mounted: " + name})

    def volume_driver_get(self):
        name = self.get_request()["Name"]
        mountpoint = os.path.join(MOUNT_DIRECTORY, name)
        if os.path.exists(mountpoint):
            self.respond({"Volume": {"Name": name, "Mountpoint": mountpoint}, "Err": ""})
        else:
            self.respond({"Err": "Not mounted: " + name})

    def volume_driver_list(self):
        volumes = os.listdir(MOUNT_DIRECTORY)
        result = [{"Name": v, "Mountpoint": os.path.join(
            MOUNT_DIRECTORY, v)} for v in volumes]
        self.respond({"Volumes": result, "Err": ""})


if __name__ == '__main__':
    read_optional_config()
    try:
        os.makedirs(MOUNT_DIRECTORY)
    except OSError as error:
        if error.errno != 17:
            raise error
    try:
        os.makedirs(PLUGIN_DIRECTORY)
    except OSError as error:
        if error.errno != 17:
            raise error
    if not is_mounted(MOUNT_DIRECTORY):
        print "Mounting Quobyte namespace in", MOUNT_DIRECTORY
        mount_all()
    print 'Starting server, use <Ctrl-C> to stop'
    UDSServer(PLUGIN_SOCKET, DockerHandler).serve_forever()
