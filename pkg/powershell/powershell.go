package powershell

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	ps "github.com/simonjanss/go-powershell"
	"github.com/simonjanss/go-powershell/middleware"
	"github.com/simonjanss/go-powershell/utils"
)

type Client struct {
	Host     string
	Username string
	Password string
	Shell    ps.Shell
	Session  middleware.Middleware
}

type Options struct {
	Host                  string
	Username              string
	Password              string
	AllowRedirection      bool
	Authentication        string
	CertificateThumbprint string
	Port                  int
	UseSSL                bool
}

// NewClient creates a new remote-client
func NewClient(shell ps.Shell, opts Options) (*Client, error) {
	// prepare remote session configuration
	config := middleware.NewSessionConfig()
	config.ComputerName = opts.Host
	if len(opts.Username) != 0 {
		config.Credential = middleware.UserPasswordCredential{
			Username: opts.Username,
			Password: opts.Password,
		}
	}

	// create a new shell by wrapping the existing one in the session middleware
	session, err := middleware.NewSession(shell, config)
	if err != nil {
		return nil, err
	}

	return &Client{
		Host:     config.ComputerName,
		Username: opts.Username,
		Password: opts.Password,
		Shell:    shell,
		Session:  session,
	}, nil
}

// CheckPath checks if the specified path exists
func (c *Client) CheckPath(path string) error {
	stdout, _, err := c.Session.Execute(fmt.Sprintf("Test-Path -path %s", path))
	if strings.HasPrefix(stdout, "False") {
		return fmt.Errorf("no such path: %s", path)
	}
	return err
}

func (c *Client) RemoveFile(path, name string) error {
	_, _, err := c.Session.Execute(fmt.Sprintf("Remove-Item -Path %s\\%s -Force", path, name))
	return err
}

func (c *Client) CreateFile(path, name string, data []byte) error {
	file, err := ioutil.TempFile(".", name)
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())

	if err := file.Close(); err != nil {
		return err
	}

	if err := ioutil.WriteFile(file.Name(), data, 0644); err != nil {
		return err
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	sess := utils.CreateRandomString(8)

	// Create new session command
	var newSessionCMD string
	if len(c.Username) == 0 {
		newSessionCMD = fmt.Sprintf("$%s = New-PSSession -ComputerName %s", sess, c.Host)
	} else {
		_, _, err := c.Shell.Execute(fmt.Sprintf("$%s = ConvertTo-SecureString -String %s -AsPlainText -Force", "pw"+sess, utils.QuoteArg(c.Password)))
		if err != nil {
			return fmt.Errorf("Could not convert password to secure string: %v", err)
		}

		_, _, err = c.Shell.Execute(fmt.Sprintf("$%s = New-Object -TypeName 'System.Management.Automation.PSCredential' -ArgumentList %s, $%s",
			"cred"+sess,
			utils.QuoteArg(c.Username),
			"pw"+sess,
		))

		if err != nil {
			return fmt.Errorf("Could not create PSCredential object : %v", err)
		}

		newSessionCMD = fmt.Sprintf("$%s = New-PSSession -ComputerName '%s' -Credential $%s", sess, c.Host, "cred"+sess)
	}

	_, _, err = c.Shell.Execute(newSessionCMD)
	if err != nil {
		os.Remove(file.Name())
		return err
	}

	// create the command for copying the file
	copyCmd := fmt.Sprintf("Copy-Item %s\\%s -Destination %s\\%s -ToSession $%s",
		wd,
		FormatFilename(file.Name()),
		path,
		FormatFilename(name),
		sess,
	)

	// execute the command
	_, _, err = c.Shell.Execute(copyCmd)
	if err != nil {
		os.Remove(file.Name())
		return err
	}

	return os.Remove(file.Name())
}

// SetupNuix will setup Nuix with the websocket-gem for ruby API
func (c *Client) SetupNuix(nuixPath string) error {
	// Set the location to nuix-path
	if err := c.setLocation(nuixPath); err != nil {
		return err
	}

	// Execute the command to install websocket-gem
	stdout, stderr, err := c.Session.Execute("jre\\bin\\java -Xmx500M -classpath lib/* org.jruby.Main --command gem install websocket --user-install")
	if err != nil {
		// Do not return the error if it contains WARNING
		if !strings.Contains(err.Error(), "WARNING") {
			return fmt.Errorf("unable to install websocket-lib to jruby - err: %v", err)
		}
	}

	// check for errors
	if stdout != "" || stderr != "" {
		// Do not return the error if it contains WARNING
		if !strings.Contains(stderr, "WARNING") {
			return fmt.Errorf("unable to install websocket-lib to jruby - stdout: %s stderr: %s", stdout, stderr)
		}
	}

	return nil
}

func (c *Client) ListGem(nuixPath string) error {
	// Set the location to nuix-path
	if err := c.setLocation(nuixPath); err != nil {
		return err
	}

	// Execute the command to install websocket-gem
	_, _, err := c.Session.Execute("jre\\bin\\java.exe -Xmx500M -classpath lib/* org.jruby.Main --command gem query -l")
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) SetEnv(variable, arg string) error {
	_, _, err := c.Session.Execute(fmt.Sprintf("$Env:%s = '%s'", variable, arg))
	return err
}

func (c *Client) Echo(arg string) (string, error) {
	stdout, stderr, err := c.Session.Execute(fmt.Sprintf("echo %s", arg))
	if err != nil {
		return "", err
	}
	if stderr != "" {
		return "", fmt.Errorf("stderr: %s", stderr)
	}
	return stdout, nil
}

func (c *Client) RemoveItem(path string) error {
	stdout, stderr, err := c.Session.Execute(fmt.Sprintf("Remove-Item -Path %s -Force", path))
	if stderr != "" {
		return fmt.Errorf("stderr: %s", stderr)
	}
	if strings.Contains(stdout, "ERROR") {
		return fmt.Errorf("stdout: %s", stdout)
	}
	return err
}

func (c *Client) RunWithCmd(path string, args ...string) error {
	// Set the location to path
	if err := c.setLocation(path); err != nil {
		return err
	}

	cmd := strings.Join(args, " ")
	stdout, stderr, err := c.Session.Execute(fmt.Sprintf("cmd.exe /c '.\\%s'", cmd))
	if stderr != "" {
		return fmt.Errorf("stderr: %s", stderr)
	}
	if strings.Contains(stdout, "ERROR") {
		return fmt.Errorf("stdout: %s", stdout)
	}
	return err
}

func (c *Client) Run(path string, args ...string) error {
	// Set the location to path
	if err := c.setLocation(path); err != nil {
		return err
	}

	cmd := strings.Join(args, " ")
	_, _, err := c.Session.Execute(".\\" + cmd)
	return err
}

func (c *Client) setLocation(path string) error {
	if err := c.CheckPath(path); err != nil {
		return err
	}

	stdout, stderr, err := c.Session.Execute(fmt.Sprintf("Set-Location %s", path))
	if err != nil {
		return fmt.Errorf("unable to set location to path: %s - %v", path, err)
	}

	if stdout != "" || stderr != "" {
		return fmt.Errorf("unable to set location to path: %s - %s %s", path, stdout, stderr)
	}
	return nil
}

func (c *Client) setLocalLocation(path string) error {
	if err := c.CheckPath(path); err != nil {
		return err
	}

	stdout, stderr, err := c.Shell.Execute(fmt.Sprintf("Set-Location %s", path))
	if err != nil {
		return fmt.Errorf("unable to set location to path: %s - %v", path, err)
	}

	if stdout != "" || stderr != "" {
		return fmt.Errorf("unable to set location to path: %s - %s %s", path, stdout, stderr)
	}
	return nil
}

func (c *Client) Close() {
	if c.Session != nil {
		c.Session.Close()
		c.Session = nil
	}
}

func (c *Client) getExitCode() (int64, error) {
	stdout, _, err := c.Session.Execute("echo $LastExitCode")
	if err != nil {
		return 2, err
	}

	stdout = strings.Replace(stdout, "\r\n", "", 1)
	return strconv.ParseInt(stdout, 10, 64)
}

func FormatFilename(filename string) string {
	if strings.Contains(filename, " ") {
		return fmt.Sprintf(`"%s"`, filename)
	}
	return filename
}

// FormatPath formats the path
// with quoutes around strings with spaces
func FormatPath(path string) string {
	if strings.Contains(path, "\\") {
		return formatPath("\\", path)
	}
	return formatPath("/", path)
}

// formatPath formats the specified path
// with quoutes around strings with spaces
func formatPath(slash, path string) string {
	// formattedPath will be returned
	var formattedPath string

	// Split path on backslashes and iterate
	splittedPath := strings.Split(path, slash)
	for i, split := range splittedPath {
		// Set quotes around the string if it contains a space
		if strings.Contains(split, " ") {
			split = fmt.Sprintf(`"%s"`, split)
		}

		// Append the string to the formattedPath variable
		formattedPath += fmt.Sprintf("%s", split)

		// Add backslashes to the path if it isn't the last dir/file
		if (i + 1) != len(splittedPath) {
			formattedPath += slash
		}
	}
	return formattedPath
}
