package ssh

import (
	"fmt"
	"io/ioutil"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

type Client struct {
	IP       string
	Port     int
	User     string
	Password string
	Cert     string
	session  *ssh.Session
	client   *ssh.Client
}

const (
	DEFAULT_TIMEOUT = 3 // second
)

func readPublicKeyFile(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(key)
}

func (client *Client) Connect() error {
	var auth []ssh.AuthMethod
	if client.Cert != "" {
		auth = []ssh.AuthMethod{readPublicKeyFile(client.Cert)}
	} else if client.Password != "" {
		auth = []ssh.AuthMethod{ssh.Password(client.Password)}
	} else {
		return fmt.Errorf("Specify authentication-method")
	}

	cfg := &ssh.ClientConfig{
		User: client.User,
		Auth: auth,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		Timeout: time.Second * DEFAULT_TIMEOUT,
	}

	connClient, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", client.IP, client.Port), cfg)
	if err != nil {
		return err
	}

	session, err := connClient.NewSession()
	if err != nil {
		if err := connClient.Close(); err != nil {
			return err
		}
		return err
	}

	client.session = session
	client.client = connClient
	return nil
}

func (client *Client) RunCmd(cmd string) error {
	out, err := client.session.CombinedOutput(cmd)
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

func (client *Client) Close() {
	if client.session != nil {
		client.session.Close()
		client.session = nil
	}

	if client.client != nil {
		client.client.Close()
		client.client = nil
	}
}
