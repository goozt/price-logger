package main

import (
	"dilogger/utils"
	"fmt"
	"os"
	"os/exec"
)

func main() {
	if utils.IsSUDO() {
		exec.Command("systemctl", "stop", "diplogger").Run()
		exec.Command("systemctl", "disable", "diplogger").Run()
		fmt.Println("DI Price Logger uninstalled")
		os.RemoveAll("/opt/dilogger")
		os.Remove("/lib/systemd/system/diplogger.service")
	} else {
		fmt.Println("Access denied. Need root access.")
	}
}
