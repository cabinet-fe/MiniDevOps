package deployer

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/crypto/ssh"
)

// CreateSSHClientConfig creates ssh.ClientConfig from ServerInfo (password or key auth)
func CreateSSHClientConfig(server ServerInfo) (*ssh.ClientConfig, error) {
	var authMethods []ssh.AuthMethod

	if server.AuthType == "key" && server.PrivateKey != "" {
		signer, err := parsePrivateKey(server.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("parse private key: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	if server.Password != "" {
		authMethods = append(authMethods, ssh.Password(server.Password))
	}

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no auth method available (password or private key required)")
	}

	config := &ssh.ClientConfig{
		User:            server.Username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	return config, nil
}

func parsePrivateKey(pem string) (ssh.Signer, error) {
	// Try without passphrase first
	signer, err := ssh.ParsePrivateKey([]byte(pem))
	if err == nil {
		return signer, nil
	}
	if !strings.Contains(err.Error(), "passphrase") {
		return nil, err
	}
	// With passphrase - not supported for now, return original error
	return nil, err
}

// ExecuteRemoteScript connects via SSH and executes a script, streaming output to logFn
func ExecuteRemoteScript(ctx context.Context, server ServerInfo, script string, logFn func(string)) error {
	config, err := CreateSSHClientConfig(server)
	if err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%d", server.Host, server.Port)
	if server.Port == 0 {
		addr = server.Host + ":22"
	}

	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return fmt.Errorf("ssh dial: %w", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("new session: %w", err)
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	if err := session.Run(script); err != nil {
		if stdout.Len() > 0 {
			for _, line := range strings.Split(strings.TrimSpace(stdout.String()), "\n") {
				logFn(line)
			}
		}
		if stderr.Len() > 0 {
			for _, line := range strings.Split(strings.TrimSpace(stderr.String()), "\n") {
				logFn("stderr: " + line)
			}
		}
		return fmt.Errorf("script execution: %w", err)
	}

	if stdout.Len() > 0 {
		for _, line := range strings.Split(strings.TrimSpace(stdout.String()), "\n") {
			logFn(line)
		}
	}
	if stderr.Len() > 0 {
		for _, line := range strings.Split(strings.TrimSpace(stderr.String()), "\n") {
			logFn("stderr: " + line)
		}
	}
	return nil
}

// buildSSHOptions returns SSH options for rsync -e and scp -o (e.g. "-o StrictHostKeyChecking=no -o Port=22")
func buildSSHOptions(server ServerInfo) string {
	opts := buildSSHOptionsSlice(server)
	return strings.Join(opts, " ")
}

// buildSSHOptionsSlice returns []string{"-o", "Opt1", "-o", "Opt2"} for scp
func buildSSHOptionsSlice(server ServerInfo) []string {
	var result []string
	result = append(result, "-o", "StrictHostKeyChecking=no")
	if server.Port > 0 && server.Port != 22 {
		result = append(result, "-o", fmt.Sprintf("Port=%d", server.Port))
	}
	if server.AuthType == "key" && server.PrivateKey != "" {
		tmpFile, err := os.CreateTemp("", "buildflow-deploy-key-*")
		if err == nil {
			tmpFile.WriteString(server.PrivateKey)
			tmpFile.Close()
			result = append(result, "-o", "IdentityFile="+tmpFile.Name())
		}
	}
	return result
}

// runAndLog executes cmd and streams output line by line to logFn
func runAndLog(cmd *exec.Cmd, logFn func(string)) error {
	if logFn == nil {
		logFn = func(string) {}
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			logFn(scanner.Text())
		}
	}()
	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		logFn(scanner.Text())
	}
	return cmd.Wait()
}
