package powershell_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/matryer/is"
	ps "github.com/simonjanss/go-powershell"
	"github.com/simonjanss/go-powershell/backend"

	"github.com/avian-digital-forensics/auto-processing/pkg/powershell"
)

func TestCheckPath(t *testing.T) {
	is := is.New(t)

	// start a local powershell process
	shell, err := ps.New(&backend.Local{})
	is.NoErr(err)
	defer shell.Close()

	// Information for the remote-server
	host := "test.avian.dk"

	// Create a new remote-client with the config and the powershell-process
	// a client holds the existing powershell-process and the remote-session
	client, err := powershell.NewClient(host, shell)
	is.NoErr(err)
	defer client.Close()

	var tt = []struct {
		path string
		err  string
	}{
		{path: "C:\\", err: ""},
		{path: "C:\\not-existing", err: "no such path: C:\\not-existing"},
	}

	for _, tc := range tt {
		err := client.CheckPath(tc.path)
		if tc.err == "" {
			is.NoErr(err)
		} else {
			is.Equal(err.Error(), tc.err)
		}
	}
}

func TestSetupNuix(t *testing.T) {
	is := is.New(t)

	// start a local powershell process
	shell, err := ps.New(&backend.Local{})
	is.NoErr(err)
	defer shell.Close()

	// Information for the remote-server
	host := "test.avian.dk" // os.Getenv("TEST_SERVER")

	// Create a new remote-client with the config and the powershell-process
	// a client holds the existing powershell-process and the remote-session
	client, err := powershell.NewClient(host, shell)
	is.NoErr(err)
	defer client.Close()

	// FormatPath formats the path if it has spaces
	nuixPath := powershell.FormatPath("C:\\Program Files\\Nuix\\Nuix 8.4")

	start := time.Now()
	//err = client.SetupNuix(nuixPath)
	err = client.ListGem(nuixPath)
	is.NoErr(err)
	fmt.Println("time elapsed:", time.Since(start))
}

func TestRun(t *testing.T) {
	is := is.New(t)

	// start a local powershell process
	shell, err := ps.New(&backend.Local{})
	is.NoErr(err)
	defer shell.Close()

	// Information for the remote-server
	host := "test.avian.dk" // os.Getenv("TEST_SERVER")

	// Create a new remote-client with the config and the powershell-process
	// a client holds the existing powershell-process and the remote-session
	client, err := powershell.NewClient(host, shell)
	is.NoErr(err)
	defer client.Close()

	// FormatPath formats the path if it has spaces
	nuixPath := powershell.FormatPath("C:\\Program Files\\Nuix\\Nuix 8.4")

	start := time.Now()
	//err = client.SetupNuix(nuixPath)
	err = client.Run(
		nuixPath,
		"nuix_console.exe",
		"-Xmx2g",
		//"-Dnuix.registry.servers=license.avian.dk",
		"-licencesourcetype server",
		"-licencetype enterprise-workstation",
		"-licencesourcelocation license.avian.dk:27443",
		"-licenceworkers 1",
		"-signout",
		"-release",
		"-interactive",
	)
	is.NoErr(err)
	fmt.Println("time elapsed:", time.Since(start))
}

func TestSetEnv(t *testing.T) {
	is := is.New(t)

	// start a local powershell process
	shell, err := ps.New(&backend.Local{})
	is.NoErr(err)
	defer shell.Close()

	// Information for the remote-server
	host := "sune.avian.dk" // os.Getenv("TEST_SERVER")

	// Create a new remote-client with the config and the powershell-process
	// a client holds the existing powershell-process and the remote-session
	client, err := powershell.NewClient(host, shell)
	is.NoErr(err)
	defer client.Close()

	var tt = []struct {
		name     string
		variable string
		arg      string
		expected string
		fail     bool
	}{
		{name: "EnvSucceed", variable: "FOO", arg: "bar", expected: "bar"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			err := client.SetEnv(tc.variable, tc.arg)
			is.NoErr(err)

			echo, err := client.Echo("$env:" + tc.variable)
			fmt.Println(echo, tc.expected)
			is.Equal(echo, tc.expected)
		})
	}

}

func TestRunWithCmd(t *testing.T) {
	is := is.New(t)

	// start a local powershell process
	shell, err := ps.New(&backend.Local{})
	is.NoErr(err)
	defer shell.Close()

	// Information for the remote-server
	host := "sune.avian.dk" // os.Getenv("TEST_SERVER")

	// Create a new remote-client with the config and the powershell-process
	// a client holds the existing powershell-process and the remote-session
	client, err := powershell.NewClientWithCredentials(host, shell, "user", "secret!")
	is.NoErr(err)
	defer client.Close()

	// FormatPath formats the path if it has spaces
	nuixPath := powershell.FormatPath("C:\\Program Files\\Nuix\\Nuix 8.4")
	scriptName := "test.rb"
	/*
		err = client.CreateFile(nuixPath, scriptName, []byte(`puts('hello')`))
		is.NoErr(err)
		defer client.RemoveFile(nuixPath, scriptName)
	*/

	xmx := "2g"
	address := "avian-server1.avian.dk"
	port := 27443
	licence := "enterprise-workstation"
	workers := 1
	err = client.RunWithCmd(
		nuixPath,
		"nuix_console.exe",
		"-Xmx"+xmx,
		"-Dnuix.registry.servers="+address,
		"-licencesourcetype", "server",
		"-licencesourcelocation", fmt.Sprintf("%s:%d", address, port),
		"-licencetype", licence,
		"-licenceworkers", fmt.Sprintf("%d", workers),
		"-signout",
		"-release",
		scriptName,
	)
	is.NoErr(err)
}

func TestCreateFile(t *testing.T) {
	is := is.New(t)

	// start a local powershell process
	shell, err := ps.New(&backend.Local{})
	is.NoErr(err)
	defer shell.Close()

	// Information for the remote-server
	host := "test.avian.dk" // os.Getenv("TEST_SERVER")

	// Create a new remote-client with the config and the powershell-process
	// a client holds the existing powershell-process and the remote-session
	client, err := powershell.NewClient(host, shell)
	is.NoErr(err)
	defer client.Close()

	// FormatPath formats the path if it has spaces
	nuixPath := powershell.FormatPath("C:\\Program Files\\Nuix\\Nuix 8.4")
	scriptName := "test .rb"
	err = client.CreateFile(nuixPath, scriptName, []byte(`puts('hello')`))
	is.NoErr(err)

	err = client.Run(
		nuixPath,
		"nuix_console.exe",
		"-Xmx2g",
		//"-Dnuix.registry.servers=license.avian.dk",
		"-licencesourcetype server",
		"-licencetype enterprise-workstation",
		"-licencesourcelocation license.avian.dk:27443",
		"-licenceworkers 1",
		"-signout",
		"-release",
		"runner.gen.rb",
	)
	is.NoErr(err)
}

func TestFormatPath(t *testing.T) {
	is := is.New(t)

	var tt = []struct {
		name     string
		path     string
		expected string
	}{
		{name: "FormatBackslash", path: "C:\\Program Files\\Nuix\\Nuix 8.4", expected: `C:\"Program Files"\Nuix\"Nuix 8.4"`},
		{name: "FormatForwardslash", path: "C:/Program Files/Nuix/Nuix 8.4", expected: `C:/"Program Files"/Nuix/"Nuix 8.4"`},
		{name: "SameBackslash", path: "C:\\Test\\Should\\Be-Same", expected: "C:\\Test\\Should\\Be-Same"},
		{name: "SameForwardslash", path: "C:/Test/Should/Be-Same", expected: "C:/Test/Should/Be-Same"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			result := powershell.FormatPath(tc.path)
			is.Equal(result, tc.expected)
		})
	}
}
