package ssh

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"golang.org/x/crypto/ssh"
)

type TunnelConfig struct {
	SSHHost         string
	SSHPort         int
	SSHUsername     string
	AuthMethod      string
	SSHPassword     string
	SSHPrivateKey   string
	SSHKeyPassphrase string
	RemoteHost      string
	RemotePort      int
}

type Tunnel struct {
	config     TunnelConfig
	client     *ssh.Client
	listener   net.Listener
	localAddr  string
	done       chan struct{}
	mu         sync.RWMutex
	alive      bool
}

func NewTunnel(config TunnelConfig) (*Tunnel, error) {
	authMethods, err := buildAuthMethods(config)
	if err != nil {
		return nil, fmt.Errorf("failed to build auth methods: %w", err)
	}

	sshConfig := &ssh.ClientConfig{
		User:            config.SSHUsername,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", config.SSHHost, config.SSHPort)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SSH server: %w", err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create local listener: %w", err)
	}

	t := &Tunnel{
		config:    config,
		client:    client,
		listener:  listener,
		localAddr: listener.Addr().String(),
		done:      make(chan struct{}),
		alive:     true,
	}

	go t.acceptLoop()

	return t, nil
}

func buildAuthMethods(config TunnelConfig) ([]ssh.AuthMethod, error) {
	var methods []ssh.AuthMethod

	if config.AuthMethod == "password" {
		if config.SSHPassword != "" {
			methods = append(methods, ssh.Password(config.SSHPassword))
		}
	} else {
		if config.SSHPrivateKey != "" {
			signer, err := parsePrivateKey(config.SSHPrivateKey, config.SSHKeyPassphrase)
			if err != nil {
				return nil, fmt.Errorf("failed to parse private key: %w", err)
			}
			methods = append(methods, ssh.PublicKeys(signer))
		}
	}

	if len(methods) == 0 {
		return nil, fmt.Errorf("no authentication method configured")
	}

	return methods, nil
}

func parsePrivateKey(key, passphrase string) (ssh.Signer, error) {
	keyBytes := []byte(key)

	if passphrase != "" {
		return ssh.ParsePrivateKeyWithPassphrase(keyBytes, []byte(passphrase))
	}

	return ssh.ParsePrivateKey(keyBytes)
}

// TestConnection tests SSH connectivity without creating a tunnel.
// It connects to the SSH server, authenticates, and immediately closes.
func TestConnection(config TunnelConfig) error {
	authMethods, err := buildAuthMethods(config)
	if err != nil {
		return fmt.Errorf("failed to build auth methods: %w", err)
	}

	sshConfig := &ssh.ClientConfig{
		User:            config.SSHUsername,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", config.SSHHost, config.SSHPort)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Close()

	// Send a keepalive to verify the connection is fully working
	_, _, err = client.SendRequest("keepalive@golang.org", true, nil)
	if err != nil {
		return fmt.Errorf("connection established but failed keepalive: %w", err)
	}

	return nil
}

func (t *Tunnel) acceptLoop() {
	for {
		select {
		case <-t.done:
			return
		default:
		}

		conn, err := t.listener.Accept()
		if err != nil {
			select {
			case <-t.done:
				return
			default:
				continue
			}
		}

		go t.handleConnection(conn)
	}
}

func (t *Tunnel) handleConnection(localConn net.Conn) {
	defer localConn.Close()

	remoteAddr := fmt.Sprintf("%s:%d", t.config.RemoteHost, t.config.RemotePort)
	log.DefaultLogger.Debug("Dialing remote address through SSH tunnel", "remoteAddr", remoteAddr)

	remoteConn, err := t.client.Dial("tcp", remoteAddr)
	if err != nil {
		log.DefaultLogger.Error("Failed to dial remote address through SSH tunnel", "remoteAddr", remoteAddr, "error", err)
		return
	}
	defer remoteConn.Close()

	log.DefaultLogger.Debug("Successfully connected to remote through SSH tunnel", "remoteAddr", remoteAddr)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		n, err := io.Copy(localConn, remoteConn)
		if err != nil {
			log.DefaultLogger.Debug("Copy from remote to local ended", "bytes", n, "error", err)
		}
	}()

	go func() {
		defer wg.Done()
		n, err := io.Copy(remoteConn, localConn)
		if err != nil {
			log.DefaultLogger.Debug("Copy from local to remote ended", "bytes", n, "error", err)
		}
	}()

	wg.Wait()
}


func (t *Tunnel) LocalAddr() string {
	return t.localAddr
}

func (t *Tunnel) IsAlive() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if !t.alive {
		return false
	}

	_, _, err := t.client.SendRequest("keepalive@golang.org", true, nil)
	return err == nil
}

func (t *Tunnel) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.alive {
		return nil
	}

	t.alive = false
	close(t.done)

	var errs []error

	if t.listener != nil {
		if err := t.listener.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if t.client != nil {
		if err := t.client.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}
