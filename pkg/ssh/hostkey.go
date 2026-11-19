package ssh

import (
	"bytes"
	"fmt"
	"net"
	"slices"
	"strings"

	"golang.org/x/crypto/ssh"
)

// UnknownHostKeyError is returned when the SSH server presents a host key that
// is not in the configured trust list. It carries the presented key so the user
// can verify its fingerprint out of band and add it to the datasource settings.
type UnknownHostKeyError struct {
	Key ssh.PublicKey
	// Configured is false when no host keys were configured at all.
	Configured bool
}

func (e *UnknownHostKeyError) Error() string {
	presented := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(e.Key)))
	fingerprint := ssh.FingerprintSHA256(e.Key)
	if !e.Configured {
		return fmt.Sprintf("no SSH host key is configured. The server presented %s (%s). "+
			"Verify this fingerprint out of band, then add the key to the SSH Host Key setting", presented, fingerprint)
	}
	return fmt.Sprintf("SSH host key mismatch: the server presented %s (%s), which matches none of the configured host keys. "+
		"The server may have been re-keyed, or the connection may be intercepted", presented, fingerprint)
}

// hostKeyVerifier holds the parsed trust list for an SSH server.
type hostKeyVerifier struct {
	keys         []ssh.PublicKey
	fingerprints []string
}

// parseHostKeys parses the SSH Host Key setting. Each non-empty line that is not
// a comment may be:
//   - a known_hosts line:      "host ssh-ed25519 AAAA..." (output of ssh-keyscan)
//   - a public key:            "ssh-ed25519 AAAA... [comment]"
//   - a SHA256 fingerprint:    "SHA256:..." (output of ssh-keygen -lf)
func parseHostKeys(input string) (*hostKeyVerifier, error) {
	v := &hostKeyVerifier{}
	for i, raw := range strings.Split(input, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "SHA256:") {
			v.fingerprints = append(v.fingerprints, strings.TrimRight(line, "="))
			continue
		}

		if key, _, _, _, err := ssh.ParseAuthorizedKey([]byte(line)); err == nil {
			v.keys = append(v.keys, key)
			continue
		}

		marker, _, key, _, _, err := ssh.ParseKnownHosts([]byte(line))
		if err != nil {
			return nil, fmt.Errorf("SSH host key line %d is not a public key, known_hosts entry or SHA256 fingerprint", i+1)
		}
		if marker != "" {
			return nil, fmt.Errorf("SSH host key line %d: known_hosts marker @%s is not supported", i+1, marker)
		}
		v.keys = append(v.keys, key)
	}
	return v, nil
}

func (v *hostKeyVerifier) empty() bool {
	return len(v.keys) == 0 && len(v.fingerprints) == 0
}

func (v *hostKeyVerifier) callback(_ string, _ net.Addr, presented ssh.PublicKey) error {
	wire := presented.Marshal()
	for _, k := range v.keys {
		if bytes.Equal(k.Marshal(), wire) {
			return nil
		}
	}
	if slices.Contains(v.fingerprints, ssh.FingerprintSHA256(presented)) {
		return nil
	}
	return &UnknownHostKeyError{Key: presented, Configured: !v.empty()}
}

// algorithms returns the host key algorithms to negotiate so the server presents
// a key of a pinned type. Returns nil (library defaults) when fingerprints are
// pinned, since their key type is unknown.
func (v *hostKeyVerifier) algorithms() []string {
	if len(v.fingerprints) > 0 || len(v.keys) == 0 {
		return nil
	}
	seen := map[string]bool{}
	var algos []string
	add := func(a string) {
		if !seen[a] {
			seen[a] = true
			algos = append(algos, a)
		}
	}
	for _, k := range v.keys {
		if k.Type() == ssh.KeyAlgoRSA {
			add(ssh.KeyAlgoRSASHA512)
			add(ssh.KeyAlgoRSASHA256)
		}
		add(k.Type())
	}
	return algos
}

// applyHostKeyPolicy sets HostKeyCallback (and HostKeyAlgorithms) on cfg
// according to the tunnel configuration.
func applyHostKeyPolicy(cfg *ssh.ClientConfig, config TunnelConfig) error {
	if config.SkipHostKeyVerify {
		// Explicit, user-controlled opt-out ("Skip host key verification").
		cfg.HostKeyCallback = ssh.InsecureIgnoreHostKey() // #nosec G106 -- user-controlled opt-in
		return nil
	}
	v, err := parseHostKeys(config.HostKeys)
	if err != nil {
		return err
	}
	cfg.HostKeyCallback = v.callback
	cfg.HostKeyAlgorithms = v.algorithms()
	return nil
}
