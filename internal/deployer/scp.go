package deployer

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type SCPDeployer struct{}

func (d *SCPDeployer) Deploy(ctx context.Context, opts DeployOptions) error {
	sshOpts := buildSSHOptionsSlice(opts.Server)
	source := strings.TrimSuffix(opts.SourceDir, "/") + "/"
	remote := fmt.Sprintf("%s@%s:%s", opts.Server.Username, opts.Server.Host, opts.RemotePath)

	args := []string{"-r"}
	args = append(args, sshOpts...)
	args = append(args, source, remote)

	cmd := exec.CommandContext(ctx, "scp", args...)
	return runAndLog(cmd, opts.Logger)
}
