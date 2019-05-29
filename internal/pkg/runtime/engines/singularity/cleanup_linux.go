// Copyright (c) 2018, Sylabs Inc. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package singularity

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/sylabs/singularity/internal/pkg/instance"
	"github.com/sylabs/singularity/internal/pkg/sylog"
)

/*
 * see https://github.com/opencontainers/runtime-spec/blob/master/runtime.md#lifecycle
 * we will run step 8/9 there
 */

// CleanupContainer cleans up the container
func (engine *EngineOperations) CleanupContainer(fatal error, status syscall.WaitStatus) error {
	sylog.Debugf("Cleanup container")

	if engine.EngineConfig.GetDeleteImage() {
		image := engine.EngineConfig.GetImage()
		sylog.Verbosef("Removing image %s", image)
		sylog.Infof("Cleaning up image...")
		if err := os.RemoveAll(image); err != nil {
			sylog.Errorf("failed to delete container image %s: %s", image, err)
		}
	}

	if engine.EngineConfig.Network != nil {
		if err := engine.EngineConfig.Network.DelNetworks(); err != nil {
			sylog.Errorf("%s", err)
		}
	}

	if engine.EngineConfig.Cgroups != nil {
		if err := engine.EngineConfig.Cgroups.Remove(); err != nil {
			sylog.Errorf("%s", err)
		}
	}

	if engine.EngineConfig.GetInstance() {
		file, err := instance.Get(engine.CommonConfig.ContainerID, instance.SingSubDir)
		if err != nil {
			return err
		}
		return file.Delete()
	}

	if engine.EngineConfig.CryptMount != "" {
		// Elevate the privilege to unmount and delete the crypt device
		uid := os.Getuid()
		err := syscall.Setresuid(uid, 0, uid)
		if err != nil {
			sylog.Debugf("Err setting suid")
			return err
		}

		err = syscall.Unmount("/usr/local/var/singularity/mnt/session/final", syscall.MNT_DETACH)
		if err != nil {
			syscall.Setresuid(uid, uid, 0)
			sylog.Debugf("Error while unmounting overlay FS: %s", err)
			return err
		}

		err = syscall.Unmount("/usr/local/var/singularity/mnt/session/rootfs", syscall.MNT_DETACH)
		if err != nil {
			syscall.Setresuid(uid, uid, 0)
			sylog.Debugf("Error while unmounting Rootfs: %s", err)
			return err
		}

		cmd := exec.Command("/sbin/cryptsetup", "luksClose", engine.EngineConfig.CryptMount)
		cmd.Dir = "/dev/mapper"
		cmd.SysProcAttr = &syscall.SysProcAttr{}
		cmd.SysProcAttr.Credential = &syscall.Credential{Uid: 0, Gid: 0}
		_, err = cmd.CombinedOutput()
		if err != nil {
			sylog.Debugf("Unable to delete the crypt device %s", err)
			syscall.Setresuid(uid, uid, 0)
			return err
		}

		// Restore the privilege
		err = syscall.Setresuid(uid, uid, 0)
	}

	return nil
}
