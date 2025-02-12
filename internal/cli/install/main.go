package main

import (
	"dilogger/internal/db"
	"dilogger/internal/utils"
	"fmt"
	"log"
	"os"
	"os/exec"
)

const dst_dir = "/opt/dilogger"

// The main function connects to a database, checks for root access, installs a service, copies files, starts and enables the service, and prints installation completion message.
func main() {
	db.ConnectDB("dilogger")
	if utils.IsSUDO() {
		installService()
		os.MkdirAll(dst_dir, 0755)
		utils.CopyFile("./app", dst_dir+"/app")
		utils.CopyFile("./reset", dst_dir+"/reset")
		utils.CopyFile("./uninstall", dst_dir+"/uninstall")
		err := exec.Command("systemctl", "start", "diplogger").Run()
		if err != nil {
			fmt.Println("Unable to start service")
			os.Exit(1)
		} else {
			exec.Command("systemctl", "enable", "diplogger").Run()
			fmt.Println("Installation complete")
		}
	} else {
		fmt.Println("Access denied. Need root access.")
	}
}

// The `installService` function creates a systemd service file for a Design Info Price Logger application.
func installService() {
	content := `[Unit]
Description=Design Info Price Logger
ConditionPathExists=` + dst_dir + `/app
After=clickhouse-server.service

[Service]
Type=simple
WorkingDirectory=` + dst_dir + `
ExecStart=` + dst_dir + `/app
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
`
	if err := os.WriteFile("/lib/systemd/system/diplogger.service", []byte(content), 0644); err != nil {
		log.Fatalln(err)
	}
}
