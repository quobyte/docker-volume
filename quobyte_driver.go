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
	tenantID     string
	configName   string
}

func newQuobyteDriver(apiURL string, username string, password string, quobyteMount string, maxFSChecks int, maxWaitTime float64, fconfigName string, fTenantID string) quobyteDriver {
	driver := quobyteDriver{
		client:       quobyte_api.NewQuobyteClient(apiURL, username, password),
		quobyteMount: quobyteMount,
		m:            &sync.Mutex{},
		maxFSChecks:  maxFSChecks,
		maxWaitTime:  maxWaitTime,
		tenantID:     fTenantID,
		configName:   fconfigName,
	}

	return driver
}

func (driver quobyteDriver) Create(request volume.Request) volume.Response {
	log.Printf("Creating volume %s\n", request.Name)
	driver.m.Lock()
	defer driver.m.Unlock()

	user, group := "root", "root"
	configurationName := driver.configName
	retryPolicy := "INTERACTIVE"
	tenantID := driver.tenantID

	if usr, ok := request.Options["user"]; ok {
		user = usr
	}
	if grp, ok := request.Options["group"]; ok {
		group = grp
	}
	if conf, ok := request.Options["configuration_name"]; ok {
		configurationName = conf
	}
	if tenant, ok := request.Options["tenant_id"]; ok {
		tenantID = tenant
	}

	if _, err := driver.client.CreateVolume(&quobyte_api.CreateVolumeRequest{
		Name:              request.Name,
		RootUserID:        user,
		RootGroupID:       group,
		ConfigurationName: configurationName,
		TenantID:          tenantID,
		Retry:             retryPolicy,
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
	// Trigger volume list refresh
	mkdErr := os.Mkdir(mPoint, 0755)
	if !os.IsExist(mkdErr) {
		// we expected ErrExist, everything else is an error
		return mkdErr
	}
	// NOTE(kaisers): Workaround for issue #9727, remove when #9628 has been implemented
	time.Sleep(1 * time.Second)

	// Verify volume is available
	_, statErr := os.Stat(mPoint)
	if statErr != nil {
		return statErr
	}
	log.Printf("Validated new volume ok: %s\n", mPoint)
	return nil
}

func (driver quobyteDriver) Remove(request volume.Request) volume.Response {
	log.Printf("Removing volume %s\n", request.Name)
	driver.m.Lock()
	defer driver.m.Unlock()

	if err := driver.client.DeleteVolumeByName(request.Name, driver.tenantID); err != nil {
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
