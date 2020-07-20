package powershell_test

import (
	"log"

	"github.com/avian-digital-forensics/auto-processing/pkg/powershell"
)

func ExampleTestConnection() {
	ps := powershell.NewClient("hostname", "username", "password")
	if err := ps.TestConnection("C:\\path"); err != nil {
		log.Fatal(err)
	}
}
