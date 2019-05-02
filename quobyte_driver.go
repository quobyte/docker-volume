package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
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

func (driver quobyteDriver) stripVolumeName(requestName string) (strippedVolumeName string, subdirPath string) {
	nameElems := strings.SplitN(requestName, "/", 2)
	if len(nameElems) == 2 {
		subDir := path.Clean(nameElems[1])
		return nameElems[0], subDir
	}
	return nameElems[0], ""
}

func (driver quobyteDriver) Create(request volume.Request) volume.Response {
	driver.m.Lock()
	defer driver.m.Unlock()

	volumeName, subDirs := driver.stripVolumeName(request.Name)
	if subDirs == "" {
		log.Printf("Creating volume %s\n", volumeName)
	} else {
		log.Printf("Creating volume %s with subdir(s) %s\n", volumeName, subDirs)
	}

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
		Name:              volumeName,
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

	mPoint := filepath.Join(driver.quobyteMount, volumeName)
	log.Printf("Validate mounting volume %s on %s\n", volumeName, mPoint)
	if err := driver.checkMountPoint(mPoint); err != nil {
		return volume.Response{Err: err.Error()}
	}

	if subDirs != "" {
		log.Printf("Creating subdir(s) %s for new volume %s\n", subDirs, volumeName)
		if csdErr := os.MkdirAll(filepath.Join(mPoint, subDirs), 0755); csdErr != nil {
			log.Printf("Unable to create subdirs in new volume: %s", csdErr)
		}
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
	driver.m.Lock()
	defer driver.m.Unlock()

	volumeName, _ := driver.stripVolumeName(request.Name)
	log.Printf("Removing volume %s\n", volumeName)
	if err := driver.client.DeleteVolumeByName(volumeName, driver.tenantID); err != nil {
		log.Println(err)
		return volume.Response{Err: err.Error()}
	}

	return volume.Response{Err: ""}
}

func (driver quobyteDriver) Mount(request volume.MountRequest) volume.Response {
	driver.m.Lock()
	defer driver.m.Unlock()
	volumeName, _ := driver.stripVolumeName(request.Name)
	mPoint := filepath.Join(driver.quobyteMount, volumeName)
	log.Printf("Mounting volume %s on %s\n", volumeName, mPoint)
	return volume.Response{Err: "", Mountpoint: mPoint}
}

func (driver quobyteDriver) Path(request volume.Request) volume.Response {
	volumeName, _ := driver.stripVolumeName(request.Name)
	return volume.Response{Mountpoint: filepath.Join(driver.quobyteMount, volumeName)}
}

func (driver quobyteDriver) Unmount(request volume.UnmountRequest) volume.Response {
	return volume.Response{}
}

func (driver quobyteDriver) Get(request volume.Request) volume.Response {
	driver.m.Lock()
	defer driver.m.Unlock()

	volumeName, _ := driver.stripVolumeName(request.Name)
	mPoint := filepath.Join(driver.quobyteMount, volumeName)

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
			volumeName, _ := driver.stripVolumeName(entry.Name())
			vols = append(vols, &volume.Volume{Name: entry.Name(), Mountpoint: filepath.Join(driver.quobyteMount, volumeName)})
		}
	}

	return volume.Response{Volumes: vols}
}

func (driver quobyteDriver) Capabilities(request volume.Request) volume.Response {
	return volume.Response{Capabilities: volume.Capability{Scope: "global"}}
}
