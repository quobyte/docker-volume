package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/docker/go-plugins-helpers/volume"
	quobyte_api "github.com/quobyte/api"
)

type quobyteDriver struct {
	client       *quobyte_api.QuobyteClient
	quobyteMount string
	m            *sync.Mutex
}

func newQuobyteDriver(apiURL string, username string, password string, quobyteMount string) quobyteDriver {
	driver := quobyteDriver{
		client:       quobyte_api.NewQuobyteClient(apiURL, username, password),
		quobyteMount: quobyteMount,
		m:            &sync.Mutex{},
	}

	return driver
}

func (driver quobyteDriver) Create(request volume.Request) volume.Response {
	log.Printf("Creating volume %s\n", request.Name)
	driver.m.Lock()
	defer driver.m.Unlock()

	user, group := "root", "root"

	if usr, ok := request.Options["user"]; ok {
		user = usr
	}

	if grp, ok := request.Options["group"]; ok {
		group = grp
	}

	if _, err := driver.client.CreateVolume(request.Name, user, group); err != nil {
		return volume.Response{Err: err.Error()}
	}

	return volume.Response{Err: ""}
}

func (driver quobyteDriver) Remove(request volume.Request) volume.Response {
	log.Printf("Removing volume %s\n", request.Name)
	driver.m.Lock()
	defer driver.m.Unlock()

	if err := driver.client.DeleteVolumeByName(request.Name); err != nil {
		return volume.Response{Err: err.Error()}
	}

	return volume.Response{Err: ""}
}

func (driver quobyteDriver) Mount(request volume.MountRequest) volume.Response {
	driver.m.Lock()
	defer driver.m.Unlock()
	mPoint := filepath.Join(driver.quobyteMount, request.Name)
	log.Printf("Mounting volume %s on %s\n", request.Name, mPoint)
	if fi, err := os.Lstat(mPoint); err != nil || !fi.IsDir() {
		return volume.Response{Err: fmt.Sprintf("%v not mounted", mPoint)}
	}

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
