package main

import (
	"dilogger/utils"
	"fmt"
	"os"
	"os/exec"
)

// The main function checks if the user has sudo privileges and stops and disables a service if so.
func main() {
	if utils.IsSUDO() {
		exec.Command("systemctl", "stop", "diplogger").Run()
		exec.Command("systemctl", "disable", "diplogger").Run()
		os.RemoveAll("/opt/dilogger")
		os.Remove("/lib/systemd/system/diplogger.service")
	} else {
		fmt.Println("Access denied. Need root access.")
	}
}
