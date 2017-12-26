package main

import (
	"dbus/org/freedesktop/login1"
	"dbus/org/freedesktop/udisks2"
	"os"
	"os/exec"
	"path"
	"pkg.deepin.io/lib/dbus"
	"sync"
)

// CDInfo sr block info
type CDInfo struct {
	Id         string
	Path       string
	MountPoint string
	Size       uint64
}

// CDInfos sr block list
type CDInfos []*CDInfo

var infoLocker sync.Mutex

func newCDInfo(blockPath dbus.ObjectPath) *CDInfo {
	block, _ := udisks2.NewBlock(udiskDBusDest, blockPath)
	defer udisks2.DestroyBlock(block)

	if block.Size.Get() == 0 || block.Id.Get() != "" {
		return nil
	}

	var info = CDInfo{
		Id:   "/dev/" + path.Base(string(blockPath)),
		Size: block.Size.Get(),
	}
	info.Path = info.Id

	return &info
}

// Get get cd info by id
func (infos CDInfos) Get(id string) *CDInfo {
	infoLocker.Lock()
	defer infoLocker.Unlock()

	for _, info := range infos {
		if info.Id == id {
			return info
		}
	}
	return nil
}

// Add add cd info, if exists, update
func (infos CDInfos) Add(info *CDInfo) CDInfos {
	if tmp := infos.Get(info.Id); tmp != nil {
		// update
		tmp = info
		return infos
	}

	infoLocker.Lock()
	doUmount(info.Path)
	dest, _ := doMount(info.Path)
	info.MountPoint = dest
	infos = append(infos, info)
	infoLocker.Unlock()
	return infos
}

// Remove remove the exists cd info
func (infos CDInfos) Remove(id string) CDInfos {
	infoLocker.Lock()
	defer infoLocker.Unlock()
	var tmp CDInfos
	for _, info := range infos {
		if info.Id == id {
			doUmount(info.MountPoint)
			continue
		}
		tmp = append(tmp, info)
	}
	return tmp
}

func doMount(src string) (string, error) {
	var dest = "/media"
	username := getCurrentUser()
	if username != "" {
		dest += "/" + username
	}
	dest += "/sr-helper-" + path.Base(src)
	os.MkdirAll(dest, 0755)
	logger.Debug("Will mount:", src, dest)
	out, err := exec.Command("mount", src, dest).CombinedOutput()
	if err != nil {
		logger.Warning("Failed to mount:", src, dest, string(out), err)
		return "", err
	}

	out, err = exec.Command("chown", "-R", username+":", dest).CombinedOutput()
	if err != nil {
		logger.Warning("Failed to chown:", username, dest, string(out), err)
	}
	return dest, err
}

func doUmount(target string) error {
	if target == "" {
		return nil
	}

	out, err := exec.Command("umount", target).CombinedOutput()
	if err != nil {
		logger.Warning("Failed to umount:", target, string(out), err)
	}
	return err
}

func getCurrentUser() string {
	self, _ := login1.NewSession("org.freedesktop.login1", "/org/freedesktop/login1/session/self")
	defer login1.DestroySession(self)
	values := self.User.Get()
	if len(values) != 2 {
		return ""
	}

	uPath := values[1].(dbus.ObjectPath)
	u, _ := login1.NewUser("org.freedesktop.login1", uPath)
	defer login1.DestroyUser(u)
	return u.Name.Get()
}
