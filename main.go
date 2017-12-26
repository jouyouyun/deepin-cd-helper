package main

import (
	"os"
	"pkg.deepin.io/lib"
	"pkg.deepin.io/lib/dbus"
	"pkg.deepin.io/lib/log"
)

const (
	dbusDest = "com.deepin.helper.CD"
	dbusPath = "/com/deepin/helper/CD"
	dbusIFC  = dbusDest
)

var logger = log.NewLogger(dbusDest)

func main() {
	if !lib.UniqueOnSystem(dbusDest) {
		logger.Error("There has a cd helper running, exit")
		return
	}

	var m = newManager()
	err := dbus.InstallOnSystem(m)
	if err != nil {
		logger.Error("Failed to install dbus:", err)
		return
	}
	dbus.DealWithUnhandledMessage()

	err = dbus.Wait()
	if err != nil {
		logger.Error("Lost dbus connection:", err)
		os.Exit(-1)
	}
	return
}
