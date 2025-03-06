package app

import (
	"bytes"
	"crypto/rand"
	"dilogger/internal/db"
	"dilogger/internal/product"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"
)

// Add all commands to root
func Start(s *db.Server) error {
	s.AddCobraCommand(NewInitCommand(s))
	s.AddCobraCommand(NewServeCommand(s.App))
	s.AddCobraCommand(NewStartCommand(s.App, false))
	s.AddCobraCommand(NewStopCommand(s.App))
	return s.App.Execute()
}

func NewInitCommand(server *db.Server) *cobra.Command {
	command := &cobra.Command{
		Use:          "init",
		Args:         cobra.ArbitraryArgs,
		Short:        "Starts the web server (default to 127.0.0.1:8090 if no domain is specified)",
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			server.NewUrlCollection()
			server.NewProductCollection()
			server.NewPriceCollection()
			InitSettings(server.App)
			AddUser(server.App, "_superusers")
			AddUser(server.App, "users")
			AddURL(server.App, "https://www.designinfo.in/wishlist/view/f6a054/")
			AddURL(server.App, "https://www.designinfo.in/wishlist/view/da0c1e/")
			product.ReloadData(server)
		},
	}
	return command
}

// Create serve command which run the server
func NewServeCommand(app core.App) *cobra.Command {
	var hideStartBanner bool
	var allowedOrigins []string
	var httpAddr string
	var httpsAddr string
	var pingBackAddr string
	command := &cobra.Command{
		Use:          "serve [domain(s)]",
		Args:         cobra.ArbitraryArgs,
		Short:        "Starts the web server (default to 127.0.0.1:8090 if no domain is specified)",
		SilenceUsage: true,
		RunE: func(command *cobra.Command, args []string) error {
			if len(args) > 0 {
				if httpAddr == "" {
					httpAddr = "0.0.0.0:80"
				}
				if httpsAddr == "" {
					httpsAddr = "0.0.0.0:443"
				}
			} else {
				if httpAddr == "" {
					httpAddr = "127.0.0.1:8090"
				}
			}

			if pingBackAddr != "" {
				confirmationBytes, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("reading confirmation bytes from stdin: %v", err)
				}
				conn, err := net.Dial("tcp", pingBackAddr)
				if err != nil {
					return fmt.Errorf("dialing confirmation address: %v", err)
				}
				defer conn.Close()
				_, err = conn.Write(confirmationBytes)
				if err != nil {
					return fmt.Errorf("writing confirmation bytes to %s: %v", pingBackAddr, err)
				}
			}

			// Warn user incase of any missing environment variables required
			hasXDG := os.Getenv("XDG_DATA_HOME") != "" &&
				os.Getenv("XDG_CONFIG_HOME") != "" &&
				os.Getenv("XDG_CACHE_HOME") != ""
			switch runtime.GOOS {
			case "windows":
				if os.Getenv("HOME") == "" && os.Getenv("USERPROFILE") == "" && !hasXDG {
					app.Logger().Warn("neither HOME nor USERPROFILE environment variables are set - please fix")
				}
			case "plan9":
				if os.Getenv("home") == "" && !hasXDG {
					app.Logger().Warn("$home environment variable is empty - please fix")
				}
			default:
				if os.Getenv("HOME") == "" && !hasXDG {
					app.Logger().Warn("$HOME environment variable is empty - please fix")
				}
			}

			err := apis.Serve(app, apis.ServeConfig{
				HttpAddr:           httpAddr,
				HttpsAddr:          httpsAddr,
				ShowStartBanner:    !hideStartBanner,
				AllowedOrigins:     allowedOrigins,
				CertificateDomains: args,
			})

			if errors.Is(err, http.ErrServerClosed) {
				return nil
			}

			return err
		},
	}

	command.PersistentFlags().StringSliceVar(
		&allowedOrigins,
		"origins",
		[]string{"*"},
		"CORS allowed domain origins list",
	)

	command.PersistentFlags().StringVar(
		&httpAddr,
		"http",
		"",
		"TCP address to listen for the HTTP server\n(if domain args are specified - default to 0.0.0.0:80, otherwise - default to 127.0.0.1:8090)",
	)

	command.PersistentFlags().StringVar(
		&httpsAddr,
		"https",
		"",
		"TCP address to listen for the HTTPS server\n(if domain args are specified - default to 0.0.0.0:443, otherwise - default to empty string, aka. no TLS)\nThe incoming HTTP traffic also will be auto redirected to the HTTPS version",
	)

	command.PersistentFlags().StringVar(
		&pingBackAddr,
		"pingback",
		"",
		"TCP address to listen for the Ping Back server",
	)

	command.PersistentFlags().BoolVar(
		&hideStartBanner,
		"hidebanner",
		false,
		"Hide Banner",
	)

	return command
}

// Creates start comand
func NewStartCommand(app core.App, showStartBanner bool) *cobra.Command {
	var allowedOrigins []string
	var httpAddr string
	var httpsAddr string

	command := &cobra.Command{
		Use:          "start [domain(s)]",
		Args:         cobra.ArbitraryArgs,
		Short:        "Starts the web server as service (default to 127.0.0.1:8090 if no domain is specified)",
		SilenceUsage: true,
		RunE: func(command *cobra.Command, args []string) error {

			// create an internal listener for testing the background process
			ln, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				return fmt.Errorf("opening listener for success confirmation: %v", err)
			}
			defer ln.Close()

			// create serve command as a background process
			cmd := exec.Command(os.Args[0], "serve", "--hidebanner", "--pingback", ln.Addr().String())
			if errors.Is(cmd.Err, exec.ErrDot) {
				cmd.Err = nil
			}

			if httpAddr != "" {
				cmd.Args = append(cmd.Args, "--http", httpAddr)
			}
			if httpsAddr != "" {
				cmd.Args = append(cmd.Args, "--https", httpsAddr)
			}
			if len(allowedOrigins) > 0 {
				cmd.Args = append(cmd.Args, slices.Insert(allowedOrigins, 0, "--origins")...)
			}

			stdinPipe, err := cmd.StdinPipe()
			if err != nil {
				return fmt.Errorf("creating stdin pipe: %v", err)
			}
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			expect := make([]byte, 32)
			_, err = rand.Read(expect)
			if err != nil {
				return fmt.Errorf("generating random confirmation bytes: %v", err)
			}

			// Send a 32 byte random string to background process
			go func() {
				_, _ = stdinPipe.Write(expect)
				stdinPipe.Close()
			}()

			// Start the command as background process
			err = cmd.Start()
			if err != nil {
				return fmt.Errorf("starting server process: %v", err)
			}

			success, exit := make(chan struct{}), make(chan error)

			// This piece of code is a goroutine that is continuously accepting incoming connections on a listener
			go func() {
				for {
					conn, err := ln.Accept()
					if err != nil {
						if !errors.Is(err, net.ErrClosed) {
							log.Println(err)
						}
						break
					}
					err = handlePingbackConn(conn, expect)
					if err == nil {
						close(success)
						break
					}
					log.Println(err)
				}
			}()

			// Goroutine that wait for error
			go func() {
				err := cmd.Wait()
				exit <- err
			}()

			// Wait for either success or exit to occur
			select {
			case <-success:
				fmt.Printf("Successfully started Server (pid=%d) - Server is running in the background\n", cmd.Process.Pid)
			case err := <-exit:
				return fmt.Errorf("server process exited with error: %v", err)
			}

			return err
		},
	}

	command.PersistentFlags().StringSliceVar(
		&allowedOrigins,
		"origins",
		[]string{"*"},
		"CORS allowed domain origins list",
	)

	command.PersistentFlags().StringVar(
		&httpAddr,
		"http",
		"",
		"TCP address to listen for the HTTP server\n(if domain args are specified - default to 0.0.0.0:80, otherwise - default to 127.0.0.1:8090)",
	)

	command.PersistentFlags().StringVar(
		&httpsAddr,
		"https",
		"",
		"TCP address to listen for the HTTPS server\n(if domain args are specified - default to 0.0.0.0:443, otherwise - default to empty string, aka. no TLS)\nThe incoming HTTP traffic also will be auto redirected to the HTTPS version",
	)

	return command
}

// Creates a command to stop a background process gracefully by sending a POST request
func NewStopCommand(app core.App) *cobra.Command {
	var port string
	command := &cobra.Command{
		Use:          "stop",
		Args:         cobra.ArbitraryArgs,
		Short:        "Stops the background process as gracefully as possible.",
		SilenceUsage: true,
		RunE: func(command *cobra.Command, args []string) error {
			address := "http://127.0.0.1:" + port

			// Get pid from file
			pid, err := os.ReadFile(filepath.Join(app.DataDir(), ".pid"))
			if err != nil {
				app.Logger().Warn("failed to stop process: process not found")
				return err
			}

			// Trigger stop api with PID as body
			postBody, _ := json.Marshal(map[string]string{"id": strings.TrimSpace(string(pid))})
			resp, err := http.Post(address+"/stop", "application/json", bytes.NewBuffer(postBody))
			if err != nil {
				app.Logger().Warn("failed using API to stop instance")
				return err
			}
			defer resp.Body.Close()
			return nil
		},
	}

	command.PersistentFlags().StringVar(
		&port,
		"port",
		"8090",
		"Local port where the server to stop is running",
	)

	return command
}

// Verify the data received back from background process
func handlePingbackConn(conn net.Conn, expect []byte) error {
	defer conn.Close()
	confirmationBytes, err := io.ReadAll(io.LimitReader(conn, 32))
	if err != nil {
		return err
	}
	if !bytes.Equal(confirmationBytes, expect) {
		return fmt.Errorf("wrong confirmation: %x", confirmationBytes)
	}
	return nil
}
