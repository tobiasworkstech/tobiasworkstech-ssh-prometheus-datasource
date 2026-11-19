package ssh

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
)

func newSigner(t *testing.T, kind string) ssh.Signer {
	t.Helper()
	var priv any
	switch kind {
	case "ed25519":
		_, k, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			t.Fatal(err)
		}
		priv = k
	case "rsa":
		k, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			t.Fatal(err)
		}
		priv = k
	}
	s, err := ssh.NewSignerFromKey(priv)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func authorizedLine(k ssh.PublicKey) string {
	return strings.TrimSpace(string(ssh.MarshalAuthorizedKey(k)))
}

func TestParseHostKeysFormats(t *testing.T) {
	key := newSigner(t, "ed25519").PublicKey()
	other := newSigner(t, "ed25519").PublicKey()

	tests := []struct {
		name    string
		input   string
		accepts bool
	}{
		{"authorized_keys line", authorizedLine(key), true},
		{"authorized_keys line with comment", authorizedLine(key) + " root@prom", true},
		{"ssh-keyscan / known_hosts line", "prom.example.com " + authorizedLine(key), true},
		{"known_hosts line with port", "[prom.example.com]:2222 " + authorizedLine(key), true},
		{"SHA256 fingerprint", ssh.FingerprintSHA256(key), true},
		{"SHA256 fingerprint with padding", ssh.FingerprintSHA256(key) + "=", true},
		{"multiple lines with comments", "# keyscan output\n\n" + authorizedLine(other) + "\nhost " + authorizedLine(key), true},
		{"different key", authorizedLine(other), false},
		{"different fingerprint", ssh.FingerprintSHA256(other), false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			v, err := parseHostKeys(tc.input)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			err = v.callback("prom.example.com:22", nil, key)
			if tc.accepts && err != nil {
				t.Fatalf("expected key to be accepted, got %v", err)
			}
			if !tc.accepts {
				var uerr *UnknownHostKeyError
				if !errors.As(err, &uerr) || !uerr.Configured {
					t.Fatalf("expected configured UnknownHostKeyError, got %v", err)
				}
			}
		})
	}
}

func TestParseHostKeysRejectsInvalid(t *testing.T) {
	for _, input := range []string{
		"not a key",
		"ssh-ed25519 !!!notbase64",
		"@cert-authority *.example.com " + authorizedLine(newSigner(t, "ed25519").PublicKey()),
	} {
		if _, err := parseHostKeys(input); err == nil {
			t.Errorf("expected error for %q", input)
		}
	}
}

func TestEmptyHostKeysReportsPresentedKey(t *testing.T) {
	key := newSigner(t, "ed25519").PublicKey()
	v, err := parseHostKeys("  \n# nothing yet\n")
	if err != nil {
		t.Fatal(err)
	}
	err = v.callback("h:22", nil, key)
	var uerr *UnknownHostKeyError
	if !errors.As(err, &uerr) || uerr.Configured {
		t.Fatalf("expected unconfigured UnknownHostKeyError, got %v", err)
	}
	if !strings.Contains(err.Error(), ssh.FingerprintSHA256(key)) || !strings.Contains(err.Error(), authorizedLine(key)) {
		t.Fatalf("error should include presented key and fingerprint: %v", err)
	}
	if strings.Contains(err.Error(), `"`) {
		t.Fatalf("error message should not contain double quotes: %v", err)
	}
}

func TestHostKeyAlgorithms(t *testing.T) {
	rsaKey := newSigner(t, "rsa").PublicKey()
	edKey := newSigner(t, "ed25519").PublicKey()

	v, _ := parseHostKeys(authorizedLine(rsaKey) + "\n" + authorizedLine(edKey))
	got := strings.Join(v.algorithms(), ",")
	want := strings.Join([]string{ssh.KeyAlgoRSASHA512, ssh.KeyAlgoRSASHA256, ssh.KeyAlgoRSA, ssh.KeyAlgoED25519}, ",")
	if got != want {
		t.Fatalf("algorithms = %s, want %s", got, want)
	}

	v, _ = parseHostKeys(ssh.FingerprintSHA256(edKey))
	if v.algorithms() != nil {
		t.Fatalf("fingerprint pins should not restrict algorithms")
	}
}

// startServer runs a minimal SSH server that accepts password "secret" and
// presents every given host key.
func startServer(t *testing.T, hostKeys ...ssh.Signer) (string, int) {
	t.Helper()
	cfg := &ssh.ServerConfig{
		PasswordCallback: func(_ ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if string(pass) == "secret" {
				return nil, nil
			}
			return nil, fmt.Errorf("denied")
		},
	}
	for _, k := range hostKeys {
		cfg.AddHostKey(k)
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go func() {
				_, chans, reqs, err := ssh.NewServerConn(c, cfg)
				if err != nil {
					c.Close()
					return
				}
				go func() {
					for r := range reqs {
						if r.WantReply {
							_ = r.Reply(true, nil)
						}
					}
				}()
				for ch := range chans {
					_ = ch.Reject(ssh.Prohibited, "no channels")
				}
			}()
		}
	}()
	host, portStr, _ := net.SplitHostPort(ln.Addr().String())
	port, _ := strconv.Atoi(portStr)
	return host, port
}

func TestTestConnectionHostKeyVerification(t *testing.T) {
	rsaHost := newSigner(t, "rsa")
	edHost := newSigner(t, "ed25519")
	host, port := startServer(t, rsaHost, edHost)

	base := TunnelConfig{SSHHost: host, SSHPort: port, SSHUsername: "u", AuthMethod: "password", SSHPassword: "secret"}

	tests := []struct {
		name    string
		mutate  func(*TunnelConfig)
		wantErr string
	}{
		{"pinned ed25519 key", func(c *TunnelConfig) { c.HostKeys = authorizedLine(edHost.PublicKey()) }, ""},
		// The server also has an ed25519 key; pinning RSA must still negotiate RSA.
		{"pinned rsa key on multi-key server", func(c *TunnelConfig) { c.HostKeys = authorizedLine(rsaHost.PublicKey()) }, ""},
		{"known_hosts line", func(c *TunnelConfig) {
			c.HostKeys = fmt.Sprintf("[%s]:%d %s", host, port, authorizedLine(edHost.PublicKey()))
		}, ""},
		{"wrong key", func(c *TunnelConfig) { c.HostKeys = authorizedLine(newSigner(t, "ed25519").PublicKey()) }, "host key mismatch"},
		{"no key configured", func(c *TunnelConfig) {}, "no SSH host key is configured"},
		{"skip verification", func(c *TunnelConfig) { c.SkipHostKeyVerify = true }, ""},
		{"malformed key", func(c *TunnelConfig) { c.HostKeys = "garbage" }, "line 1"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := base
			tc.mutate(&cfg)
			err := TestConnection(cfg)
			if tc.wantErr == "" && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tc.wantErr)) {
				t.Fatalf("expected error containing %q, got %v", tc.wantErr, err)
			}
		})
	}
}
