package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// The `Environment` type in Go represents configuration settings for a host, port, database, username, and password.
type Environment struct {
	Host     string
	Port     string
	Database string
	Username string
	Password string
}

// The `getEnv` function retrieves the value of an environment variable or returns a fallback value if the variable is not set.
func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

// The GetEnvironment function loads environment variables and returns a struct containing database connection details with default values if not set.
func GetEnvironment() Environment {
	godotenv.Load()
	return Environment{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "9000"),
		Database: getEnv("DATABASE", "default"),
		Username: getEnv("DB_USERNAME", "default"),
		Password: getEnv("DB_PASSWORD", "password"),
	}
}

// The GetHTML function retrieves the HTML content of a given URL and returns it as an io.ReadCloser along with any errors encountered.
func GetHTML(url string) (io.ReadCloser, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// The IsSUDO function checks if the current process is running with root privileges.
func IsSUDO() bool {
	stdout, err := exec.Command("ps", "-o", "user=", "-p", strconv.Itoa(os.Getpid())).Output()
	if err != nil {
		fmt.Println("Access denied")
		os.Exit(1)
	}
	return strings.TrimSpace(string(stdout)) == "root"
}

// The function `CopyFile` copies a file from the source path to the destination path in Go.
func CopyFile(src, dst string) {
	err := exec.Command("cp", "-rf", src, dst).Run()
	if err != nil {
		fmt.Println(err)
		fmt.Println("Unable to copy files")
		os.Exit(1)
	}
}
