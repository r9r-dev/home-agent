package services

import (
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSHTestResult contains the result of an SSH connection test
type SSHTestResult struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	LatencyMs int64  `json:"latency_ms,omitempty"`
}

// TestSSHConnection tests connectivity to an SSH server
func TestSSHConnection(host string, port int, username, authType, authValue string) SSHTestResult {
	start := time.Now()

	var authMethod ssh.AuthMethod
	if authType == "password" {
		authMethod = ssh.Password(authValue)
	} else {
		// Parse SSH private key
		signer, err := ssh.ParsePrivateKey([]byte(authValue))
		if err != nil {
			return SSHTestResult{
				Success: false,
				Message: fmt.Sprintf("Cle SSH invalide: %v", err),
			}
		}
		authMethod = ssh.PublicKeys(signer)
	}

	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{authMethod},
		// Note: InsecureIgnoreHostKey is acceptable for internal use
		// For production with external hosts, consider implementing known_hosts verification
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return SSHTestResult{
			Success: false,
			Message: fmt.Sprintf("Connexion echouee: %v", err),
		}
	}
	defer client.Close()

	latency := time.Since(start).Milliseconds()

	return SSHTestResult{
		Success:   true,
		Message:   "Connexion reussie",
		LatencyMs: latency,
	}
}
