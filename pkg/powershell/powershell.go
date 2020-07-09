package powershell

import (
	"fmt"
	"strconv"
	"strings"

	ps "github.com/simonjanss/go-powershell"
	"github.com/simonjanss/go-powershell/middleware"
)

type Client struct {
	Host        string
	ProgramPath string
	Shell       ps.Shell
	Session     middleware.Middleware
}

// NewClient creates a new remote-client
func NewClient(host, programPath string, shell ps.Shell) (*Client, error) {
	// prepare remote session configuration
	//config := &middleware.SessionConfig{ComputerName: host}
	config := middleware.NewSessionConfig()
	config.ComputerName = host

	// create a new shell by wrapping the existing one in the session middleware
	session, err := middleware.NewSession(shell, config)
	if err != nil {
		return nil, err
	}

	return &Client{
		Host:        host,
		ProgramPath: programPath,
		Shell:       shell,
		Session:     session,
	}, nil
}

func (c *Client) Run(cfg string) error {
	// Format the paths
	newCfg, cfgPath := c.formatPaths(cfg)

	// Copy the config to the remote-server
	if err := c.copyConfig(cfg, cfgPath); err != nil {
		return fmt.Errorf("Failed to copy config to remote-server: %v", err)
	}

	// Remove the config from the remote-server
	defer c.rmConfig(cfgPath)

	if err := c.execute(newCfg); err != nil {
		return fmt.Errorf("Failed to execute program with config %s : %v", newCfg, err)
	}

	exitcode, err := c.getExitCode()
	if err != nil {
		return fmt.Errorf("Failed to get exitcode: %v", err)
	}

	if exitcode != 0 {
		return fmt.Errorf("Program exited with code: %d", exitcode)
	}

	return nil
}

func (c *Client) Close() {
	if c.Session != nil {
		c.Session.Close()
		c.Session = nil
	}
}

func (c *Client) formatPaths(cfg string) (string, string) {
	newCfg := strings.ReplaceAll(cfg, "./configs/", "")
	netPath := strings.Replace(c.ProgramPath, ":", "$", 1)
	cfgPath := fmt.Sprintf("//%s/%s/%s", c.Host, netPath, newCfg)
	return newCfg, cfgPath
}

func (c *Client) copyConfig(cfg, cfgPath string) error {
	_, _, err := c.Shell.Execute(fmt.Sprintf("Copy-Item -Path %s -Destination %s", cfg, cfgPath))
	return err
}

func (c *Client) rmConfig(cfgPath string) {
	_, _, err := c.Shell.Execute(fmt.Sprintf("Remove-Item -Path %s", cfgPath))
	if err != nil {
		fmt.Println("Couldn't remove config:", err)
	}
}

// Execute the program with the specified config
func (c *Client) execute(cfg string) error {
	// everything run via the session is run on the remote machine
	_, _, err := c.Session.Execute(fmt.Sprintf("cd %s", c.ProgramPath))
	if err != nil {
		return err
	}

	// everything run via the session is run on the remote machine
	_, _, err = c.Session.Execute(fmt.Sprintf(".\\auto-processing.exe --cfg=%s", cfg))
	return err
}

func (c *Client) getExitCode() (int64, error) {
	stdout, _, err := c.Session.Execute("echo $LastExitCode")
	if err != nil {
		return 2, err
	}

	stdout = strings.Replace(stdout, "\r\n", "", 1)
	return strconv.ParseInt(stdout, 10, 64)
}
