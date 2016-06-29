package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

func readOptionalConfig() {
	mountQuobytePath = os.Getenv("MOUNT_QUOBYTE_PATH")
	if len(mountQuobyteOptions) == 0 {
		mountQuobytePath = "/run/docker/quobyte/mnt"
	}
	mountQuobyteOptions = os.Getenv("MOUNT_QUOBYTE_OPTIONS")
	if len(mountQuobyteOptions) == 0 {
		mountQuobyteOptions = "-o user_xattr"
	}
}

func readMandatoryConfig() {
	qmgmtUser = getMandatoryEnv("QUOBYTE_API_USER")
	qmgmtPassword = getMandatoryEnv("QUOBYTE_API_PASSWORD")
	quobyteAPIURL = getMandatoryEnv("QUOBYTE_API_URL")
	quobyteRegistry = getMandatoryEnv("QUOBYTE_REGISTRY")
}

func getMandatoryEnv(name string) string {
	env := os.Getenv(name)
	if len(env) < 0 {
		log.Fatalf("Please set %s in environment\n", name)
	}

	return env
}

func isMounted(mountPath string) bool {
	content, err := ioutil.ReadFile("/proc/mounts")
	if err != nil {
		log.Println(err)
	}
	for _, mount := range strings.Split(string(content), "\n") {
		splitted := strings.Split(mount, " ")
		if len(splitted) < 2 {
			continue
		}

		if splitted[1] == mountPath {
			log.Printf("Found Mountpoint: %s\n", mountPath)
			return true
		}
	}

	return false
}

func mountAll() {
	cmdStr := fmt.Sprintf("mount %s -t quobyte %s %s", mountQuobyteOptions, fmt.Sprintf("%s/", quobyteRegistry), mountQuobytePath)
	if out, err := exec.Command("/bin/sh", "-c", cmdStr).CombinedOutput(); err != nil {
		log.Fatalln(string(out))
	}
}
