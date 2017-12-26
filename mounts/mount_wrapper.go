package main

import (
	"fmt"
	"pkg.deepin.io/lib/dbus"
)

type cdInfo struct {
	Id         string
	Path       string
	MountPoint string
	Size       uint64
}

const (
	cdDBusDest = "com.deepin.helper.CD"
	cdDBusPath = "/com/deepin/helper/CD"
)

func getCDList(conn *dbus.Conn) {
	cd := conn.Object(cdDBusDest, cdDBusPath)
	var r dbus.Variant
	err := cd.Call("org.freedesktop.DBus.Properties.Get",
		0, cdDBusDest, "List").Store(&r)
	if err != nil {
		fmt.Println("Failed to get list:", err)
		return
	}
	fmt.Println("Signature:", r.Signature().String())
	if r.Signature().String() != "a(ssst)" {
		fmt.Println("Invalid list data:", r.Signature().String())
		return
	}

	values := r.Value().([][]interface{})
	for _, value := range values {
		fmt.Println(value[0].(string), value[1].(string), value[2].(string), value[3].(uint64))
	}
}

func main() {
	conn, err := dbus.SystemBus()
	if err != nil {
		fmt.Println("Failed to connection bus:", err)
		return
	}
	defer conn.Close()

	addRule := "type=signal" + ",path=" + cdDBusPath + ",member=Added"
	removeRule := "type=signal" + ",path=" + cdDBusPath + ",member=Removed"
	conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, addRule)
	conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, removeRule)

	sigChan := conn.Signal()
	for s := range sigChan {
		fmt.Println("Signal:", s.Name)
		if s.Name == cdDBusDest+".Added" {
			// added
			fmt.Println("Body len:", len(s.Body))
			if len(s.Body) != 1 {
				continue
			}
			v := s.Body[0].([]interface{})
			fmt.Println("Added:", v[0].(string), v[1].(string), v[2].(string), v[3].(uint64))
		} else if s.Name == cdDBusDest+".Removed" {
			// removed
			fmt.Println("Body len:", len(s.Body))
			if len(s.Body) != 1 {
				continue
			}
			v := s.Body[0].([]interface{})
			fmt.Println("Removed:", v[0].(string), v[1].(string), v[2].(string), v[3].(uint64))
		}
	}
}
