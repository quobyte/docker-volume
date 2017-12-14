package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/docker/go-plugins-helpers/volume"
	quobyte_api "github.com/quobyte/api"
)

type quobyteDriver struct {
	client       *quobyte_api.QuobyteClient
	quobyteMount string
	m            *sync.Mutex
	maxFSChecks  int
	maxWaitTime  float64
}

func newQuobyteDriver(apiURL string, username string, password string, quobyteMount string, maxFSChecks int, maxWaitTime float64, configName string, tenantId string) quobyteDriver {
	driver := quobyteDriver{
		client:       quobyte_api.NewQuobyteClient(apiURL, username, password),
		quobyteMount: quobyteMount,
		m:            &sync.Mutex{},
		maxFSChecks:  maxFSChecks,
		maxWaitTime:  maxWaitTime,
	}

	return driver
}

func (driver quobyteDriver) Create(request volume.Request) volume.Response {
	log.Printf("Creating volume %s\n", request.Name)
	driver.m.Lock()
	defer driver.m.Unlock()

	user, group := "root", "root"
	configuration_name := "BASE"
	retry_policy := "INTERACTIVE"
	tenant_id := "default"

	if usr, ok := request.Options["user"]; ok {
		user = usr
	}

	if grp, ok := request.Options["group"]; ok {
		group = grp
	}

	if conf, ok := request.Options["configuration_name"]; ok {
		configuration_name = conf
	}

	if tenant, ok := request.Options["tenant_id"]; ok {
		tenant_id = tenant
	}

	if _, err := driver.client.CreateVolume(&quobyte_api.CreateVolumeRequest{
		Name:              request.Name,
		RootUserID:        user,
		RootGroupID:       group,
		ConfigurationName: configuration_name,
		TenantID:          tenant_id,
		Retry:             retry_policy,
	}); err != nil {
		log.Println(err)

		if !strings.Contains(err.Error(), "ENTITY_EXISTS_ALREADY/POSIX_ERROR_NONE") {
			return volume.Response{Err: err.Error()}
		}
	}

	mPoint := filepath.Join(driver.quobyteMount, request.Name)
	log.Printf("Validate mounting volume %s on %s\n", request.Name, mPoint)
	if err := driver.checkMountPoint(mPoint); err != nil {
		return volume.Response{Err: err.Error()}
	}

	return volume.Response{Err: ""}
}

func (driver quobyteDriver) checkMountPoint(mPoint string) error {
	start := time.Now()

	backoff := 1
	tries := 0
	var mount_error error
	for tries <= driver.maxFSChecks {
		mount_error = nil
		if fi, err := os.Lstat(mPoint); err != nil || !fi.IsDir() {
			log.Printf("Unsuccessful Filesystem Check for %s after %d tries", mPoint, tries)
			mount_error = err
		} else {
			return nil
		}

		time.Sleep(time.Duration(backoff) * time.Second)
		if time.Since(start).Seconds() > driver.maxWaitTime {
			log.Printf("Abort checking mount point do to time out after %f\n", driver.maxWaitTime)
			return mount_error
		}

		backoff *= 2
	}

	return mount_error
}

func (driver quobyteDriver) Remove(request volume.Request) volume.Response {
	log.Printf("Removing volume %s\n", request.Name)
	driver.m.Lock()
	defer driver.m.Unlock()

	if err := driver.client.DeleteVolumeByName(request.Name, ""); err != nil {
		log.Println(err)
		return volume.Response{Err: err.Error()}
	}

	return volume.Response{Err: ""}
}

func (driver quobyteDriver) Mount(request volume.MountRequest) volume.Response {
	driver.m.Lock()
	defer driver.m.Unlock()
	mPoint := filepath.Join(driver.quobyteMount, request.Name)
	log.Printf("Mounting volume %s on %s\n", request.Name, mPoint)
	return volume.Response{Err: "", Mountpoint: mPoint}
}

func (driver quobyteDriver) Path(request volume.Request) volume.Response {
	return volume.Response{Mountpoint: filepath.Join(driver.quobyteMount, request.Name)}
}

func (driver quobyteDriver) Unmount(request volume.UnmountRequest) volume.Response {
	return volume.Response{}
}

func (driver quobyteDriver) Get(request volume.Request) volume.Response {
	driver.m.Lock()
	defer driver.m.Unlock()

	mPoint := filepath.Join(driver.quobyteMount, request.Name)

	if fi, err := os.Lstat(mPoint); err != nil || !fi.IsDir() {
		log.Println(err)
		return volume.Response{Err: fmt.Sprintf("%v not mounted", mPoint)}
	}

	return volume.Response{Volume: &volume.Volume{Name: request.Name, Mountpoint: mPoint}}
}

func (driver quobyteDriver) List(request volume.Request) volume.Response {
	driver.m.Lock()
	defer driver.m.Unlock()

	var vols []*volume.Volume
	files, err := ioutil.ReadDir(driver.quobyteMount)
	if err != nil {
		log.Println(err)
		return volume.Response{Err: err.Error()}
	}

	for _, entry := range files {
		if entry.IsDir() {
			vols = append(vols, &volume.Volume{Name: entry.Name(), Mountpoint: filepath.Join(driver.quobyteMount, entry.Name())})
		}
	}

	return volume.Response{Volumes: vols}
}

func (driver quobyteDriver) Capabilities(request volume.Request) volume.Response {
	return volume.Response{Capabilities: volume.Capability{Scope: "global"}}
}
