// Copyright 2024 gofrp (https://github.com/gofrp)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gssh

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"os/user"
	"path/filepath"

	"golang.org/x/crypto/ssh"

	"github.com/gofrp/tiny-frpc/pkg/util"
	"github.com/gofrp/tiny-frpc/pkg/util/log"
)

type TunnelClient struct {
	localAddr string
	sshServer string
	command   string

	sshConn *ssh.Client
	ln      net.Listener

	authMethod ssh.AuthMethod
}

// find existing key in sshDir by priority, return first found path or empty if none
func findExistingKey(sshDir string) (string, error) {
	candidates := []string{
		filepath.Join(sshDir, "id_ed25519"),
		filepath.Join(sshDir, "id_rsa"),
	}
	for _, p := range candidates {
		_, err := os.Stat(p)
		if err == nil {
			return p, nil
		}
		if !os.IsNotExist(err) {
			return "", fmt.Errorf("failed to stat key file %v: %v", p, err)
		}
	}
	return "", nil
}

// find existing key by priority (ed25519 > rsa), auto-generate ed25519 if none found
func getDefaultPrivateKeyPath() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %v", err)
	}
	sshDir := filepath.Join(usr.HomeDir, ".ssh")

	if p, err := findExistingKey(sshDir); err != nil {
		return "", err
	} else if p != "" {
		return p, nil
	}

	// none found, auto-generate ed25519 key to ~/.ssh/
	defaultKeyPath := filepath.Join(sshDir, "id_ed25519")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create ssh directory: %v", err)
	}
	log.Infof("no existing ssh key found, generating ed25519 key at: [%v]", defaultKeyPath)
	if err := generateED25519Key(defaultKeyPath); err != nil {
		return "", fmt.Errorf("failed to generate private key: %v", err)
	}
	return defaultKeyPath, nil
}

// generate ed25519 key pair without passphrase protection
func generateED25519Key(path string) error {
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("failed to generate ed25519 key: %v", err)
	}

	privPEM, err := ssh.MarshalPrivateKey(privKey, "")
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %v", err)
	}
	if err := os.WriteFile(path, pem.EncodeToMemory(privPEM), 0600); err != nil {
		return fmt.Errorf("failed to write private key file: %v", err)
	}

	pubSSH, err := ssh.NewPublicKey(pubKey)
	if err != nil {
		return fmt.Errorf("failed to create public key: %v", err)
	}
	if err := os.WriteFile(path+".pub", ssh.MarshalAuthorizedKey(pubSSH), 0644); err != nil {
		return fmt.Errorf("failed to write public key file: %v", err)
	}

	log.Infof("ed25519 key generated at: [%v]", path)
	return nil
}

func publicKeyAuthFunc(kPath string) (ssh.AuthMethod, error) {
	key, err := os.ReadFile(kPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read private key: %v", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key: %v", err)
	}

	return ssh.PublicKeys(signer), nil
}

func NewTunnelClient(localAddr string, sshServer string, command string) (*TunnelClient, error) {
	privateKeyPath, err := getDefaultPrivateKeyPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get default private key path: %v", err)
	}

	log.Infof("get ssh private key file: [%v] to communicate with frps by ssh protocol", privateKeyPath)

	authMethod, err := publicKeyAuthFunc(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to generate auth method: %v", err)
	}

	return &TunnelClient{
		localAddr:  localAddr,
		sshServer:  sshServer,
		command:    command,
		authMethod: authMethod,
	}, nil
}

func (c *TunnelClient) Start() error {
	config := &ssh.ClientConfig{
		User:            "v0",
		Auth:            []ssh.AuthMethod{c.authMethod},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	conn, err := ssh.Dial("tcp", c.sshServer, config)
	if err != nil {
		return err
	}
	c.sshConn = conn

	l, err := conn.Listen("tcp", "0.0.0.0:80")
	if err != nil {
		return err
	}
	c.ln = l

	session, err := c.sshConn.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	err = session.Start(c.command)
	if err != nil {
		return err
	}

	log.Infof("session start cmd [%v] success", c.command)

	c.serveListener()
	return nil
}

func (c *TunnelClient) Close() {
	if c.sshConn != nil {
		_ = c.sshConn.Close()
	}
	if c.ln != nil {
		_ = c.ln.Close()
	}
}

func (c *TunnelClient) serveListener() {
	for {
		conn, err := c.ln.Accept()
		if err != nil {
			log.Errorf("ssh tunnel cient accept error: %v", err)
			return
		}

		log.Infof("accept a new connection. remote: %v, local: %v", conn.RemoteAddr().String(), conn.LocalAddr().String())

		go c.hanldeConn(conn)
	}
}

func (c *TunnelClient) hanldeConn(conn net.Conn) {
	defer conn.Close()
	local, err := net.Dial("tcp", c.localAddr)
	if err != nil {
		log.Errorf("ssh tunnel client dial %v error: %v", c.localAddr, err)
		return
	}
	_, _ = util.Join(local, conn)
}
