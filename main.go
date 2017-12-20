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

func getEnvWithDefault(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {

	maxFSChecksDefaultStr := getEnvWithDefault("MAX_FS_CHECKS", "5")
	maxFSChecksDefault, _ := strconv.Atoi(maxFSChecksDefaultStr)
	maxWaitTimeDefaultStr := getEnvWithDefault("MAX_WAIT_TIME", "64")
	maxWaitTimeDefault, _ := strconv.ParseFloat(maxWaitTimeDefaultStr, 64)
	quobyteAPIURLDefault := getEnvWithDefault("QUOBYTE_API_URL", "http://localhost:7860")
	quobyteAPIPasswordDefault := getEnvWithDefault("QUOBYTE_API_PASSWORD", "quobyte")
	quobyteAPIUserDefault := getEnvWithDefault("QUOBYTE_API_USER", "admin")
	quobyteMountPathDefault := getEnvWithDefault("QUOBYTE_MOUNT_PATH", "/run/docker/quobyte/mnt")
	quobyteMountOptionsDefault := getEnvWithDefault("QUOBYTE_MOUNT_OPTIONS", "-o user_xattr")
	quobyteRegistryDefault := getEnvWithDefault("QUOBYTE_REGISTRY", "localhost:7861")
	quobyteTenantIDDefault := getEnvWithDefault("QUOBYTE_TENANT_ID", "NO-DEFAULT-CHANGE-ME")
	quobyteVolConfigNameDefault := getEnvWithDefault("QUOBYTE_VOLUME_CONFIG_NAME", "BASE")
	socketGroupDefault := getEnvWithDefault("SOCKET_GROUP", "root")

	maxFSChecks := flag.Int("max-fs-checks", maxFSChecksDefault,
		"Maximimum number of filesystem checks when a Volume is created before returning an error")
	maxWaitTime := flag.Float64("max-wait-time", maxWaitTimeDefault,
		"Maximimum wait time for filesystem checks to complete when a Volume is created before returning an error")
	quobyteAPIUser := flag.String("user", quobyteAPIUserDefault, "User to connect to the Quobyte API server")
	quobyteAPIPassword := flag.String("password", quobyteAPIPasswordDefault,
		"Password for the user to connect to the Quobyte API server")
	quobyteAPIURL := flag.String("api", quobyteAPIURLDefault,
		"URL to the API server(s) in the form http(s)://host[:port][,host:port] or SRV record name")
	quobyteMountPath := flag.String("path", quobyteMountPathDefault, "Path where Quobyte is mounted on the host")
	quobyteMountOptions := flag.String("options", quobyteMountOptionsDefault,
		"Fuse options to be used when Quobyte is mounted")
	quobyteRegistry := flag.String("registry", quobyteRegistryDefault,
		"URL to the registry server(s) in the form of host[:port][,host:port] or SRV record name")
	quobyteTenantID := flag.String("tenant_id", quobyteTenantIDDefault,
		"Id of the Quobyte tenant in whose domain the operation takes place")
	quobyteVolConfigName := flag.String("configuration_name", quobyteVolConfigNameDefault,
		"Name of the volume configuration of new volumes")
	socketGroup := flag.String("group", socketGroupDefault, "Group to create the unix socket")
	showVersion := flag.Bool("version", false, "Shows version string")

	flag.Parse()

	log.Printf("\nVariables read:\n"+
		"MAX_FS_CHECKS: %v\nMAX_WAIT_TIME: %v\nSOCKET_GROUP: %s\n"+
		"QUOBYTE_API_URL: %s\nQUOBYTE_API_USER: %s\nQUOBYTE_MOUNT_PATH:"+
		" %s\nQUOBYTE_MOUNT_OPTIONS: %s\nQUOBYTE_REGISTRY: %s\nQUOBYTE_TENANT_ID: "+
		" %s\nQUOBYTE_VOLUME_CONFIG_NAME: %s\n", *maxFSChecks, *maxWaitTime,
		*socketGroup, *quobyteAPIURL, *quobyteAPIUser,
		*quobyteMountPath, *quobyteMountOptions, *quobyteRegistry, *quobyteTenantID,
		*quobyteVolConfigName)

	if *showVersion {
		log.Printf("\nVersion: %s - Revision: %s\n", version, revision)
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

	qDriver := newQuobyteDriver(*quobyteAPIURL, *quobyteAPIUser, *quobyteAPIPassword,
		*quobyteMountPath, *maxFSChecks, *maxWaitTime, *quobyteVolConfigName, *quobyteTenantID)
	handler := volume.NewHandler(qDriver)

	log.Println(handler.ServeUnix(*socketGroup, quobyteID))
}
