package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"
)

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

		if !strings.HasPrefix(splitted[0], "quobyte") {
			continue
		}

		if splitted[1] == mountPath {
			log.Printf("Found Mountpoint: %s\n", mountPath)
			return true
		}
	}

	return false
}

func mountAll(mountQuobyteOptions, quobyteRegistry, mountQuobytePath string) {
	cmdStr := fmt.Sprintf("mount %s -t quobyte %s %s", mountQuobyteOptions, fmt.Sprintf("%s/", quobyteRegistry), mountQuobytePath)
	if out, err := exec.Command("/bin/sh", "-c", cmdStr).CombinedOutput(); err != nil {
		log.Fatalln(string(out))
	}
}
