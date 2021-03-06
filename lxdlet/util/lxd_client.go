/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"fmt"

	"github.com/lxc/lxd/client"
	"github.com/lxc/lxd/shared/api"
)

type lxdDaemon struct {
	s            lxd.ContainerServer
	path         string
	images       []api.Image
	networks     []api.Network
	storagePools []api.StoragePool
}

func NewLxdClient(path string) (*lxdDaemon, error) {
	// Connect to the LXD daemon
	s, err := lxd.ConnectLXDUnix(fmt.Sprintf("%s/unix.socket", path), nil)
	if err != nil {
		return nil, err
	}

	// Setup our internal struct
	d := &lxdDaemon{s: s, path: path}
	return d, nil
}

func (d *lxdDaemon) GetInfo() (*api.Server, error) {
	info, _, err := d.s.GetServer()
	if err != nil {
		return nil, err
	}

	return info, nil
}

func (d *lxdDaemon) GetContainers() ([]api.Container, error) {
	// Containers
	containers, err := d.s.GetContainers()
	if err != nil {
		return nil, err
	}

	return containers, nil
}

func (d *lxdDaemon) CreateContainer(name string, image string, wait bool) (*lxd.Operation, error) {
	// Container creation request
	// TODO(kjackal): Kubernetes probably pulled the image localy first
	req := api.ContainersPost{
		Name: name,
		Source: api.ContainerSource{
			Type:     "image",
			Alias:    image,
			Server:   "https://images.linuxcontainers.org",
			Protocol: "simplestreams",
		},
	}

	// Get LXD to create the container (background operation)
	op, err := d.s.CreateContainer(req)
	if err != nil {
		return nil, err
	}

	if wait {
		err = op.Wait()
		if err != nil {
			return nil, err
		}
	}

	return op, nil
}

func (d *lxdDaemon) GetContainer(name string) (*api.Container, error) {
	// Containers
	container, _, err := d.s.GetContainer(name)
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (d *lxdDaemon) GetContainerState(name string) (*api.ContainerState, error) {
	// Containers
	container, _, err := d.s.GetContainerState(name)
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (d *lxdDaemon) StartContainer(name string, wait bool) (*lxd.Operation, error) {
	reqState := api.ContainerStatePut{
		Action:  "start",
		Timeout: -1,
	}

	op, err := d.s.UpdateContainerState(name, reqState, "")
	if err != nil {
		return nil, err
	}

	if wait {
		err = op.Wait()
		if err != nil {
			return nil, err
		}
	}

	return op, nil
}

func (d *lxdDaemon) StopContainer(name string, wait bool) (*lxd.Operation, error) {
	reqState := api.ContainerStatePut{
		Action:  "stop",
		Timeout: -1,
		Force:   true,
	}

	op, err := d.s.UpdateContainerState(name, reqState, "")
	if err != nil {
		return nil, err
	}

	if wait {
		err = op.Wait()
		if err != nil {
			return nil, err
		}
	}

	return op, nil
}

func (d *lxdDaemon) DeleteContainer(name string, wait bool) (*lxd.Operation, error) {
	op, err := d.s.DeleteContainer(name)
	if err != nil {
		return nil, err
	}

	if wait {
		err = op.Wait()
		if err != nil {
			return nil, err
		}
	}

	return op, nil
}

func (d *lxdDaemon) ListImages() ([]api.Image, error) {
	// Containers
	// Connect to the remote SimpleStreams server
	remote, err := lxd.ConnectSimpleStreams("https://images.linuxcontainers.org", nil)
	if err != nil {
		return nil, err
	}

	images, err := remote.GetImages()
	if err != nil {
		return nil, err
	}

	return images, nil
}

func (d *lxdDaemon) PullImage(imageName string, wait bool) (*lxd.RemoteOperation, error) {
	// Connect to the remote SimpleStreams server
	remote, err := lxd.ConnectSimpleStreams("https://images.linuxcontainers.org", nil)
	if err != nil {
		return nil, err
	}

	alias, _, err := remote.GetImageAlias(imageName)
	if err != nil {
		return nil, err
	}

	// Get the image information
	image, _, err := remote.GetImage(alias.Target)
	if err != nil {
		return nil, err
	}

	args := &lxd.ImageCopyArgs{
		Aliases: image.Aliases,
	}
	// Ask LXD to copy the image from the remote server
	op, err := d.s.CopyImage(remote, *image, args)
	if err != nil {
		return nil, err
	}

	// And wait for it to finish
	if wait {
		err = op.Wait()
		if err != nil {
			return nil, err
		}
	}

	return op, nil
}

func (d *lxdDaemon) DeleteImage(imageName string, wait bool) (*lxd.Operation, error) {
	// Get the image information
	alias, _, err := d.s.GetImageAlias(imageName)
	if err != nil {
		return nil, err
	}

	// Get the image information
	image, _, err := d.s.GetImage(alias.Target)
	if err != nil {
		return nil, err
	}

	op, err := d.s.DeleteImage(image.Fingerprint)
	if err != nil {
		return nil, err
	}
	// And wait for it to finish
	if wait {
		err = op.Wait()
		if err != nil {
			return nil, err
		}
	}

	return op, nil

}

func (d *lxdDaemon) StatImage(imageName string) (*api.Image, error) {
	// Get the image information
	alias, _, err := d.s.GetImageAlias(imageName)
	if err != nil {
		return nil, err
	}

	// Get the image information
	image, _, err := d.s.GetImage(alias.Target)
	if err != nil {
		return nil, err
	}

	return image, nil
}
