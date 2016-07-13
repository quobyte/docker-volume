package main

import (
	"flag"
	"log"
	"os"

	"github.com/docker/go-plugins-helpers/volume"
)

const quobyteID string = "quobyte"

func main() {
	quobyteMountPath := flag.String("path", "/run/docker/quobyte/mnt", "Path where Quobyte is mounted on the host")
	quobyteMountOptions := flag.String("options", "-o user_xattr", "Fuse options to be used when Quobyte is mounted")

	quobyteUser := flag.String("user", "root", "User to connect to the Quobyte API server")
	quobytePassword := flag.String("password", "quobyte", "Password for the user to connect to the Quobyte API server")
	quobyteAPIURL := flag.String("api", "localhost:7860", "URL to the API server(s) in the form host[:port][,host:port] or SRV record name")
	quobyteRegistry := flag.String("registry", "localhost:7861", "URL to the registry server(s) in the form of host[:port][,host:port] or SRV record name")

	group := flag.String("group", "root", "Group to create the unix socket")
	flag.Parse()

	if err := os.MkdirAll(*quobyteMountPath, 0555); err != nil {
		log.Println(err.Error())
	}

	if !isMounted(*quobyteMountPath) {
		log.Printf("Mounting Quobyte namespace in %s", *quobyteMountPath)
		mountAll(*quobyteMountOptions, *quobyteRegistry, *quobyteMountPath)
	}

	qDriver := newQuobyteDriver(*quobyteAPIURL, *quobyteUser, *quobytePassword, *quobyteMountPath)
	handler := volume.NewHandler(qDriver)
	log.Println(handler.ServeUnix(*group, quobyteID))
}
