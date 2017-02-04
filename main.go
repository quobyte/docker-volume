package main

import (
	"flag"
	"log"
	"os"
	"os/user"
	"strconv"

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
	quobyteAPIURL := flag.String("api", "http://localhost:7860", "URL to the API server(s) in the form http(s)://host[:port][,host:port] or SRV record name")
	quobyteRegistry := flag.String("registry", "localhost:7861", "URL to the registry server(s) in the form of host[:port][,host:port] or SRV record name")

	group := flag.String("group", "root", "Group to create the unix socket")
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

	qDriver := newQuobyteDriver(*quobyteAPIURL, *quobyteUser, *quobytePassword, *quobyteMountPath)
	handler := volume.NewHandler(qDriver)

	g, err := user.LookupGroup(*group)
	if err != nil {
		log.Fatalln(err)
	}

	gid, err := strconv.Atoi(g.Gid)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(handler.ServeUnix(quobyteID, gid))
}
