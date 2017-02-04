package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os/exec"
	"strings"
)

func validateAPIURL(apiURL string) error {
	url, err := url.Parse(apiURL)
	if err != nil {
		return err
	}
	if url.Scheme == "" {
		return fmt.Errorf("Scheme is no set in URL: %s", apiURL)
	}

	return nil
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
