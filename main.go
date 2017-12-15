package main

import (
	"flag"
	"log"
	"os"
	"strconv"

	"github.com/docker/go-plugins-helpers/volume"
)

const (
	quobyteID string = "quobyte"
)

var (
	version  string
	revision string
)

func main() {
	showVersion := flag.Bool("version", false, "Shows version string")

	flag.Parse()

	maxFSChecks, _ := strconv.Atoi(os.Getenv("MAX_FS_CHECKS"))
	maxWaitTime, _ := strconv.ParseFloat(os.Getenv("MAX_WAIT_TIME"), 64)
	socketGroup := os.Getenv("SOCKET_GROUP")

	quobyteAPIURL := os.Getenv("QUOBYTE_API_URL")
	quobyteAPIPassword := os.Getenv("QUOBYTE_API_PASSWORD")
	quobyteAPIUser := os.Getenv("QUOBYTE_API_USER")
	quobyteMountPath := os.Getenv("QUOBYTE_MOUNT_PATH")
	quobyteMountOptions := os.Getenv("QUOBYTE_MOUNT_OPTIONS")
	quobyteRegistry := os.Getenv("QUOBYTE_REGISTRY")
	quobyteTenantID := os.Getenv("QUOBYTE_TENANT_ID")
	quobyteVolConfigName := os.Getenv("QUOBYTE_VOLUME_CONFIG_NAME")
	log.Printf("\nVariables read from environment:\n"+
		"MAX_FS_CHECKS: %v\nMAX_WAIT_TIME: %v\nSOCKET_GROUP: %s\n"+
		"QUOBYTE_API_URL: %s\nQUOBYTE_API_USER: %s\nQUOBYTE_MOUNT_PATH:"+
		" %s\nQUOBYTE_MOUNT_OPTIONS: %s\nQUOBYTE_REGISTRY: %s\nQUOBYTE_TENANT_ID: "+
		" %s\nQUOBYTE_VOLUME_CONFIG_NAME: %s\n", maxFSChecks, maxWaitTime,
		socketGroup, quobyteAPIURL, quobyteAPIUser,
		quobyteMountPath, quobyteMountOptions, quobyteRegistry, quobyteTenantID,
		quobyteVolConfigName)

	if *showVersion {
		log.Printf("\nVersion: %s - Revision: %s\n", version, revision)
		return
	}

	if err := validateAPIURL(quobyteAPIURL); err != nil {
		log.Fatalln(err)
	}

	if err := os.MkdirAll(quobyteMountPath, 0555); err != nil {
		log.Println(err.Error())
	}

	if !isMounted(quobyteMountPath) {
		log.Printf("Mounting Quobyte namespace in %s", quobyteMountPath)
		mountAll(quobyteMountOptions, quobyteRegistry, quobyteMountPath)
	}

	qDriver := newQuobyteDriver(quobyteAPIURL, quobyteAPIUser, quobyteAPIPassword,
		quobyteMountPath, maxFSChecks, maxWaitTime, quobyteVolConfigName, quobyteTenantID)
	handler := volume.NewHandler(qDriver)

	log.Println(handler.ServeUnix(socketGroup, quobyteID))
}
