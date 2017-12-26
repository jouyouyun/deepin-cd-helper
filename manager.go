package main

import (
	"dbus/org/freedesktop/udisks2"
	"path"
	"pkg.deepin.io/lib/dbus"
	"strings"
	"sync"
)

const (
	udiskDBusDest = "org.freedesktop.UDisks2"
)

// Manager cd helper
type Manager struct {
	objManager *udisks2.ObjectManager
	srBlocks   []*udisks2.Block
	srLocker   sync.Mutex

	List CDInfos

	Added   func(CDInfo)
	Removed func(CDInfo)
}

func newManager() *Manager {
	var m = new(Manager)

	m.objManager, _ = udisks2.NewObjectManager(udiskDBusDest, "/org/freedesktop/UDisks2")
	m.init()
	return m
}

func (m *Manager) init() {
	srPaths := m.getSrPaths()
	logger.Info("Sr list:", srPaths)
	for _, p := range srPaths {
		m.addBlockMonitor(p)

		info := newCDInfo(p)
		if info == nil {
			continue
		}
		logger.Debug("Add sr info:", info)
		m.List = m.List.Add(info)
	}

	m.objManager.ConnectInterfacesAdded(func(p dbus.ObjectPath, detail map[string]map[string]dbus.Variant) {
		if !strings.Contains(string(p), "/org/freedesktop/UDisks2/Block/sr") {
			return
		}

		m.addBlockMonitor(p)

		info := newCDInfo(p)
		if info == nil {
			return
		}
		logger.Debug("Add sr info:", info)
		m.List = m.List.Add(info)
		dbus.Emit(m, "Added", *info)
	})

	m.objManager.ConnectInterfacesRemoved(func(p dbus.ObjectPath, ifcs []string) {
		if !strings.Contains(string(p), "/org/freedesktop/UDisks2/Block/sr") {
			return
		}

		m.removeBlock(p)
		id := "/dev/" + path.Base(string(p))
		info := m.List.Get(id)
		if info == nil {
			return
		}
		logger.Debug("Remove sr info:", info)
		m.List = m.List.Remove(id)
		dbus.Emit(m, "Removed", *info)
	})
}

func (m *Manager) addBlockMonitor(blockPath dbus.ObjectPath) {
	block, _ := udisks2.NewBlock(udiskDBusDest, blockPath)
	logger.Debug("[addBlockmonitor] info:", blockPath, block.Size.Get(), block.Id.Get())
	block.Size.ConnectChanged(func() {
		logger.Debug("Block size changed:", block.Path, block.Size.Get())
		if block.Size.Get() != 0 {
			if block.Id.Get() != "" {
				return
			}
			info := newCDInfo(blockPath)
			if info == nil {
				return
			}
			logger.Debug("Add sr info:", info)
			m.List = m.List.Add(info)
			dbus.Emit(m, "Added", *info)
		} else {
			// remove
			id := "/dev/" + path.Base(string(blockPath))
			info := m.List.Get(id)
			if info == nil {
				return
			}
			logger.Debug("Remove sr info:", info)
			m.List = m.List.Remove(id)
			dbus.Emit(m, "Removed", *info)
		}
	})
	m.addSrBlock(block)
}

func (m *Manager) getSrPaths() []dbus.ObjectPath {
	objs, err := m.objManager.GetManagedObjects()
	if err != nil {
		logger.Warning("Failed to get manager objects:", err)
		return nil
	}

	var srPaths []dbus.ObjectPath
	for p, v := range objs {
		_, ok := v["org.freedesktop.UDisks2.Block"]
		if !ok {
			continue
		}
		logger.Info("Block obj:", p)

		if !strings.Contains(string(p), "/org/freedesktop/UDisks2/block_devices/sr") {
			continue
		}

		srPaths = append(srPaths, p)
	}
	return srPaths
}

func (m *Manager) addSrBlock(block *udisks2.Block) {
	m.srLocker.Lock()
	defer m.srLocker.Unlock()
	for _, b := range m.srBlocks {
		if b.Path == block.Path {
			return
		}
	}
	m.srBlocks = append(m.srBlocks, block)
}

func (m *Manager) removeBlock(blockPath dbus.ObjectPath) {
	m.srLocker.Lock()
	defer m.srLocker.Unlock()
	var blocks []*udisks2.Block
	for _, b := range m.srBlocks {
		if b.Path == blockPath {
			udisks2.DestroyBlock(b)
			continue
		}
		blocks = append(blocks, b)
	}
	m.srBlocks = blocks
}

func (*Manager) GetDBusInfo() dbus.DBusInfo {
	return dbus.DBusInfo{
		Dest:       dbusDest,
		ObjectPath: dbusPath,
		Interface:  dbusIFC,
	}
}
