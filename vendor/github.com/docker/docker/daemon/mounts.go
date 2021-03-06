package daemon

import (
	"strings"

	"github.com/docker/docker/container"
	derr "github.com/docker/docker/errors"
	volumestore "github.com/docker/docker/volume/store"
)

func (daemon *Daemon) prepareMountPoints(container *container.Container) error {
	for _, config := range container.MountPoints {
		if err := daemon.lazyInitializeVolume(container.ID, config); err != nil {
			return err
		}
	}
	return nil
}

func (daemon *Daemon) removeMountPoints(container *container.Container, rm bool) error {
	var rmErrors []string
	for _, m := range container.MountPoints {
		if m.Volume == nil {
			continue
		}
		daemon.volumes.Dereference(m.Volume, container.ID)
		if rm {
			// Do not remove named mountpoints
			// these are mountpoints specified like `docker run -v <name>:/foo`
			if m.Named {
				continue
			}
			err := daemon.volumes.Remove(m.Volume)
			// Ignore volume in use errors because having this
			// volume being referenced by other container is
			// not an error, but an implementation detail.
			// This prevents docker from logging "ERROR: Volume in use"
			// where there is another container using the volume.
			if err != nil && !volumestore.IsInUse(err) {
				rmErrors = append(rmErrors, err.Error())
			}
		}
	}
	if len(rmErrors) > 0 {
		return derr.ErrorCodeRemovingVolume.WithArgs(strings.Join(rmErrors, "\n"))
	}
	return nil
}
