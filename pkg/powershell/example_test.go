package powershell_test

import (
	ps "github.com/simonjanss/go-powershell"
	"github.com/simonjanss/go-powershell/backend"

	"github.com/avian-digital-forensics/auto-processing/pkg/powershell"
)

func ExampleRun() {
	// start a local powershell process
	shell, err := ps.New(&backend.Local{})
	if err != nil {
		panic(err)
	}
	defer shell.Close()

	// Information for the remote-server
	host := "server.avian.dk"
	programPath := "/path/to/program"

	// Create a new remote-client with the config and the powershell-process
	// a client holds the existing powershell-process and the remote-session
	client, err := powershell.NewClient(host, programPath, shell)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	if err := client.Run("config.yml"); err != nil {
		panic(err)
	}
}
