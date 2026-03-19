package deployer

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type RsyncDeployer struct{}

func (d *RsyncDeployer) Deploy(ctx context.Context, opts DeployOptions) error {
	sshOpts := buildSSHOptions(opts.Server)
	source := strings.TrimSuffix(opts.SourceDir, "/") + "/"
	remote := fmt.Sprintf("%s@%s:%s", opts.Server.Username, opts.Server.Host, opts.RemotePath)

	var sshCmd string
	if opts.Server.Password != "" && opts.Server.AuthType != "key" {
		sshCmd = fmt.Sprintf("sshpass -p %q ssh %s", opts.Server.Password, sshOpts)
	} else {
		sshCmd = "ssh " + sshOpts
	}
	args := []string{
		"-avz",
		"--delete",
		"-e", sshCmd,
		source,
		remote,
	}

	cmd := exec.CommandContext(ctx, "rsync", args...)
	return runAndLog(cmd, opts.Logger)
}
