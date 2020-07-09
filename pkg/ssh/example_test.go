package ssh_test

import (
	"fmt"

	"github.com/avian-digital-forensics/auto-processing/pkg/ssh"
)

func ExampleConnect() {
	client := &ssh.Client{
		IP:       "localhost",
		Port:     22,
		User:     "root",
		Password: "password",
		// Cert: "/path/to/cert",
	}

	if err := client.Connect(); err != nil {
		fmt.Println(err)
	}

	if err := client.RunCmd("ls /etc"); err != nil {
		fmt.Println(err)
	}

	client.Close()
}
