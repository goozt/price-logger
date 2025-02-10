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

type Environment struct {
	Host     string
	Port     string
	Database string
	Username string
	Password string
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func GetEnvironment() Environment {
	godotenv.Load()
	return Environment{
		Host:     getEnv("HOST", "localhost"),
		Port:     getEnv("POST", "9000"),
		Database: getEnv("DATABASE", "default"),
		Username: getEnv("DB_USERNAME", "default"),
		Password: getEnv("DB_PASSWORD", "password"),
	}
}

func GetHTML(url string) (io.ReadCloser, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func IsSUDO() bool {
	stdout, err := exec.Command("ps", "-o", "user=", "-p", strconv.Itoa(os.Getpid())).Output()
	if err != nil {
		fmt.Println("Access denied")
		os.Exit(1)
	}
	return strings.TrimSpace(string(stdout)) == "root"
}

func CopyFile(src, dst string) {
	err := exec.Command("cp", "-rf", src, dst).Run()
	if err != nil {
		fmt.Println(err)
		fmt.Println("Unable to copy files")
		os.Exit(1)
	}
}
