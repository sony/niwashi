// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2026 Sony Group Corporation

package helpers

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

const sshServerImage = "lscr.io/linuxserver/openssh-server:latest"
const sshUsername = "nwsuser"

// SSHServer represents a running SSH server container for E2E testing.
type SSHServer struct {
	Host           string
	Port           int
	Username       string
	PrivateKeyPath string
	KnownHostsPath string
	containerName  string
}

// StartSSHServer starts a Docker-based SSH server and registers cleanup via t.Cleanup.
// Requires Docker to be available on the host.
func StartSSHServer(t *testing.T, dir string) *SSHServer {
	t.Helper()

	port, err := getFreePort()
	if err != nil {
		t.Fatalf("failed to get free port: %v", err)
	}

	keyDir := dir
	privateKeyPath := filepath.Join(keyDir, "id_ed25519")
	knownHostsPath := filepath.Join(keyDir, "known_hosts")

	if out, err := exec.Command("ssh-keygen", "-t", "ed25519", "-f", privateKeyPath, "-N", "").CombinedOutput(); err != nil {
		t.Fatalf("failed to generate SSH key: %v\n%s", err, out)
	}

	containerName := fmt.Sprintf("niwashi-e2e-ssh-%d", port)
	pubKeyPath := privateKeyPath + ".pub"

	out, err := exec.Command("docker", "run", "-d",
		"--name", containerName,
		"-e", "PUID=1000",
		"-e", "PGID=1000",
		"-e", fmt.Sprintf("USER_NAME=%s", sshUsername),
		"-e", "PUBLIC_KEY_FILE=/pubkey",
		"-v", fmt.Sprintf("%s:/pubkey:ro", pubKeyPath),
		"-p", fmt.Sprintf("%d:2222", port),
		sshServerImage,
	).CombinedOutput()
	if err != nil {
		t.Fatalf("failed to start SSH server container: %v\n%s", err, out)
	}

	srv := &SSHServer{
		Host:           "127.0.0.1",
		Port:           port,
		Username:       sshUsername,
		PrivateKeyPath: privateKeyPath,
		KnownHostsPath: knownHostsPath,
		containerName:  containerName,
	}

	t.Cleanup(func() {
		exec.Command("docker", "rm", "-f", containerName).Run() //nolint:errcheck
	})

	if err := srv.waitReady(30 * time.Second); err != nil {
		t.Fatalf("SSH server did not become ready: %v", err)
	}

	scanOut, err := exec.Command("ssh-keyscan", "-p", fmt.Sprintf("%d", port), "127.0.0.1").Output()
	if err != nil {
		t.Fatalf("ssh-keyscan failed: %v", err)
	}
	if err := os.WriteFile(knownHostsPath, scanOut, 0600); err != nil {
		t.Fatalf("failed to write known_hosts: %v", err)
	}

	return srv
}

// WriteInstancesYAML writes instances.yaml to dir with this server's connection details.
// The generated file can be referenced from target.yaml via external-instance.from=file provisioner.
func (s *SSHServer) WriteInstancesYAML(t *testing.T, dir string) string {
	t.Helper()

	content := fmt.Sprintf(`inst-1:
  connection:
    ssh:
      address:
        host: %s
        port: %d
        user: %s
      auth:
        privateKeyPath: %s
      hostKey:
        knownHostsPath: %s
`, s.Host, s.Port, s.Username, s.PrivateKeyPath, s.KnownHostsPath)

	path := filepath.Join(dir, "instances.yaml")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create directory %s: %v", dir, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write instances.yaml: %v", err)
	}
	return path
}

func (s *SSHServer) waitReady(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	addr := net.JoinHostPort(s.Host, strconv.Itoa(s.Port))
	for time.Now().Before(deadline) {
		out, err := exec.Command("ssh-keyscan", "-p", fmt.Sprintf("%d", s.Port), s.Host).Output()
		if err == nil && strings.Contains(string(out), "ssh") {
			return nil
		}
		// TCP が繋がるまで待ってからリトライ
		if conn, err := net.DialTimeout("tcp", addr, time.Second); err == nil {
			conn.Close()
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("timeout waiting for SSH server on %s", addr)
}

func getFreePort() (int, error) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
