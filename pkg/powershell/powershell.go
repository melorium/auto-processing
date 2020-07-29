package powershell

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"syscall"
)

// Client holds information -
// needed for a powershell-connection
type Client struct {
	Host     string
	Username string
	Password string
	LogFile  *os.File
}

// NewClient creates a new client with information -
// needed for a powershell-connection
func NewClient(host, username, password string, logFile *os.File) *Client {
	return &Client{
		Host:     host,
		Username: username,
		Password: password,
		LogFile:  logFile,
	}
}

// AutoProcessing runs the auto-processing script
func (c *Client) AutoProcessing(archive, path, cfg string) error {
	return handleError(c.execute(c.autoProcessing(archive, path, cfg)))
}

// TestConnection test the connection and if the path is specified
func (c *Client) TestConnection(path string) error {
	return handleError(c.execute(c.testConnection(path)))
}

func handleError(err error) error {
	if exiterr, ok := err.(*exec.ExitError); ok {
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			if status.ExitStatus() == 40 {
				return fmt.Errorf("%v : Cannot access remote-computer", err)
			} else if status.ExitStatus() == 50 {
				return fmt.Errorf("%v : Cannot access file-path", err)
			}
		}
		return fmt.Errorf("%v : Unknown error", err)
	}
	return nil
}

// execute the script
func (c *Client) execute(script string) error {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("powershell.exe", script)
	} else if runtime.GOOS == "darwin" {
		cmd = exec.Command("pwsh", script)
	} else if runtime.GOOS == "linux" {
		return errors.New("linux powershell not available yet - will be available soon")
	} else {
		return fmt.Errorf("%s is not avaibable yet", runtime.GOOS)
	}

	cmd.Stdout = io.MultiWriter(os.Stdout, c.LogFile)
	cmd.Stderr = io.MultiWriter(os.Stderr, c.LogFile)

	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Wait()
}
