package deployer

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

type SCPDeployer struct{}

func (d *SCPDeployer) Deploy(ctx context.Context, opts DeployOptions) error {
	sshOpts := buildSSHOptionsSlice(opts.Server)
	source := strings.TrimSuffix(opts.SourceDir, string(filepath.Separator)) + string(filepath.Separator)
	remotePath := normalizeRemotePath(opts.Server, opts.RemotePath)
	remote := fmt.Sprintf("%s@%s:%s", opts.Server.Username, opts.Server.Host, remotePath)

	args := []string{"-r"}
	args = append(args, sshOpts...)
	args = append(args, source, remote)

	cmd := exec.CommandContext(ctx, "scp", args...)
	return runAndLog(cmd, opts.Logger)
}
