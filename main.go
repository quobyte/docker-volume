package main

import (
	"log"
	"os"

	"github.com/docker/go-plugins-helpers/volume"
)

const quobyteID string = "quobyte"

// Mandatory configuration
var qmgmtUser string
var qmgmtPassword string
var quobyteAPIURL string
var quobyteRegistry string

// Optional configuration
var mountQuobytePath string
var mountQuobyteOptions string

func main() {
	readMandatoryConfig()
	readOptionalConfig()

	if err := os.MkdirAll(mountQuobytePath, 0555); err != nil {
		log.Println(err.Error())
	}

	if !isMounted(mountQuobytePath) {
		log.Printf("Mounting Quobyte namespace in %s", mountQuobytePath)
		mountAll()
	}

	qDriver := newQuobyteDriver(quobyteAPIURL, qmgmtUser, qmgmtPassword)
	handler := volume.NewHandler(qDriver)
	log.Println(handler.ServeUnix("root", quobyteID))
}
