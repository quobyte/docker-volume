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
import urlparse
import BaseHTTPServer
import json
import os, os.path
import socket
import sys
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

def mount_all(path):
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
    return qmgmt("volume create " + name + " root root " + volume_config + " 777")

def volume_delete(name):
    return qmgmt("volume delete -f " + name)

def volume_exists(name):
    return qmgmt("volume resolve " + name)

def is_mounted(mountpath):
    f = open('/proc/mounts')
    for l in f:
        if l.split()[1] == mountpath:
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
        BaseHTTPServer.HTTPServer.__init__(self, server_address, RequestHandlerClass)

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
        ret = self.socket.accept()
        return ret[0], 'uds'

class DockerHandler(BaseHTTPRequestHandler):
  mount_paths = {}

  def get_request(self):
      length = int(self.headers['content-length'])
      return json.loads(self.rfile.read(length))

  def respond(self, msg):
      self.send_response(200)
      self.send_header("Content-type", "application/vnd.docker.plugins.v1+json")
      self.end_headers()
      print "Responding with", json.dumps(msg)
      self.wfile.write(json.dumps(msg))
 
  def do_POST(self):
    if self.path == "/Plugin.Activate":
      self.respond({"Implements": ["VolumeDriver"]})

    elif self.path == "/VolumeDriver.Create":
      request = self.get_request()
      print request

      is_persistent = False
      volume_config = DEFAULT_VOLUME_CONFIGURATION

      if 'Opts' in request and request['Opts'] and 'persistent' in request['Opts']:
        valuestr = request['Opts']['persistent']
        is_persistent = eval(valuestr, {"__builtins__":None},{}) == True
      if 'Opts' in request and request['Opts'] and 'volume_configuration' in request['Opts']:
        volume_config = request['Opts']['volume_configuration']

      volume_create(request["Name"], volume_config)

      mountpoint = os.path.join(MOUNT_DIRECTORY, request["Name"])
      while True:
          if os.path.exists(mountpoint):
              break
          print "Waiting for", mountpoint
          time.sleep(1)

      self.respond({"Err": ""})

    elif self.path == "/VolumeDriver.Remove":
      request = self.get_request()
      print request
      if not volume_exists(request["Name"]):
          self.respond({"Err": ""})
          return
      if volume_delete(request["Name"]):
          self.respond({"Err": ""})
      else:
          self.respond({"Err": "Could not delete " + request["Name"]})          

    elif self.path == "/VolumeDriver.Path" or self.path == "/VolumeDriver.Mount":
      request = self.get_request()
      print request
      mountpoint = os.path.join(MOUNT_DIRECTORY, request["Name"])
      if os.path.exists(mountpoint):
          self.respond({"Err": "", "Mountpoint": mountpoint})
      else:
          self.respond({"Err": "Not mounted: " + request["Name"]}) 

    elif self.path == "/VolumeDriver.Get":
      request = self.get_request()
      print request
      mountpoint = os.path.join(MOUNT_DIRECTORY, request["Name"])
      if os.path.exists(mountpoint):
          self.respond(
            "Volume": {"Name": request["Name"], "Mountpoint": mountpoint},
            "Err": "")
      else:
          self.respond({"Err": "Not mounted: " + request["Name"]}) 

    elif self.path == "/VolumeDriver.Unmount":
      request = self.get_request()
      print request
      self.respond({"Err": ""})

    elif self.path == "/VolumeDriver.List":
      request = self.get_request()
      print request
      volumes = os.listdir(MOUNT_DIRECTORY)
      result = [{"Name": v, "Mountpoint": os.path.join(MOUNT_DIRECTORY, v)} for v in volumes]
      self.respond({"Volumes": result, "Err": ""})

    else:
      print "Unknown API operation:", self.path
      self.respond({"Err": "Unknown API operation: " + self.path})

if __name__ == '__main__':
     try:
         os.makedirs(MOUNT_DIRECTORY)
     except OSError, e:
         if e.errno != 17:
             raise e
     try:
         os.makedirs(PLUGIN_DIRECTORY)
     except OSError, e:
         if e.errno != 17:
             raise e
     if not is_mounted(MOUNT_DIRECTORY):
         print "Mounting Quobyte namespace in", MOUNT_DIRECTORY 
         mount_all(MOUNT_DIRECTORY)
     server = UDSServer(PLUGIN_SOCKET, DockerHandler)
     print 'Starting server, use <Ctrl-C> to stop'
     server.serve_forever()
