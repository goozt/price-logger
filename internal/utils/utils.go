package utils

import (
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// The `GetEnv` function retrieves the value of an environment variable or returns a fallback value if the variable is not set.
func GetEnv(key string, fallback ...string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if len(value) == 0 {
		if len(fallback) < 1 {
			log.Fatalf("error: Environtment variable '%s' not found", key)
		}
		return fallback[0]
	}
	return value
}

// The IsSUDO function checks if the current process is running with root privileges.
func IsSUDO() bool {
	stdout, err := exec.Command("ps", "-o", "user=", "-p", strconv.Itoa(os.Getpid())).Output()
	if err != nil {
		log.Println("access denied", err)
		os.Exit(1)
	}
	return strings.TrimSpace(string(stdout)) == "root"
}

// The function `CopyFile` copies a file from the source path to the destination path in Go.
func CopyFile(src, dst string) {
	err := exec.Command("cp", "-rf", src, dst).Run()
	if err != nil {
		log.Println("unable to copy files:", err)
		os.Exit(1)
	}
}
