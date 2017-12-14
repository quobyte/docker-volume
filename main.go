package main

import (
	"flag"
	"log"
	"os"

	"github.com/docker/go-plugins-helpers/volume"
)

const quobyteID string = "quobyte"

var (
	version  string
	revision string
)

func main() {
	quobyteMountPath := flag.String("path", "/run/docker/quobyte/mnt", "Path where Quobyte is mounted on the host")
	quobyteMountOptions := flag.String("options", "-o user_xattr", "Fuse options to be used when Quobyte is mounted")

	quobyteUser := flag.String("user", "root", "User to connect to the Quobyte API server")
	quobytePassword := flag.String("password", "quobyte", "Password for the user to connect to the Quobyte API server")
	quobyteConfigName := flag.String("configuration_name", "BASE", "Name of the volume configuration of new volumes")
	quobyteAPIURL := flag.String("api", "http://localhost:7860", "URL to the API server(s) in the form http(s)://host[:port][,host:port] or SRV record name")
	quobyteRegistry := flag.String("registry", "localhost:7861", "URL to the registry server(s) in the form of host[:port][,host:port] or SRV record name")
	quobyteTenantId := flag.String("tenant_id", "no default", "Id of the Quobyte tenant in whose domain the operation takes place")

	group := flag.String("group", "root", "Group to create the unix socket")
	maxWaitTime := flag.Float64("max-wait-time", 30, "Maximimum wait time for filesystem checks to complete when a Volume is created before returning an error")
	maxFSChecks := flag.Int("max-fs-checks", 5, "Maximimum number of filesystem checks when a Volume is created before returning an error")
	showVersion := flag.Bool("version", false, "Shows version string")
	flag.Parse()

	if *showVersion {
		log.Printf("Version: %s - Revision: %s\n", version, revision)
		return
	}

	if err := validateAPIURL(*quobyteAPIURL); err != nil {
		log.Fatalln(err)
	}

	if err := os.MkdirAll(*quobyteMountPath, 0555); err != nil {
		log.Println(err.Error())
	}

	if !isMounted(*quobyteMountPath) {
		log.Printf("Mounting Quobyte namespace in %s", *quobyteMountPath)
		mountAll(*quobyteMountOptions, *quobyteRegistry, *quobyteMountPath)
	}

	qDriver := newQuobyteDriver(*quobyteAPIURL, *quobyteUser, *quobytePassword, *quobyteMountPath, *maxFSChecks, *maxWaitTime, *quobyteConfigName, *quobyteTenantId)
	handler := volume.NewHandler(qDriver)

	log.Println(handler.ServeUnix(*group, quobyteID))
}
