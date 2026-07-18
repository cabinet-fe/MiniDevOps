package service

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	resourcemodel "bedrock/internal/resource/model"
)

// CLILookup is the narrow resource-domain surface AgentService depends on
// (aligned with engine.SecretResolver). Implemented by resource CLIService.
type CLILookup interface {
	FindByKey(key string) (*resourcemodel.CliRuntimeDefinition, error)
}

// ResolveBinary returns absolute path for a CLI binary if installed.
func ResolveBinary(cli *resourcemodel.CliRuntimeDefinition) (string, error) {
	if cli.InstalledPath != "" {
		if _, err := os.Stat(cli.InstalledPath); err == nil {
			return cli.InstalledPath, nil
		}
	}
	return exec.LookPath(cli.BinaryName)
}

// BuildRuntimeEnv injects API base / env templates without overwriting CLI login state files.
func BuildRuntimeEnv(cli *resourcemodel.CliRuntimeDefinition, apiBase string, extra map[string]string) []string {
	env := os.Environ()
	if apiBase != "" && cli.APIBaseEnv != "" {
		env = append(env, cli.APIBaseEnv+"="+apiBase)
	}
	for k, v := range extra {
		if strings.TrimSpace(k) == "" {
			continue
		}
		env = append(env, k+"="+v)
	}
	if cli.InstalledPath != "" {
		dir := filepath.Dir(cli.InstalledPath)
		env = append(env, "PATH="+dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	}
	return env
}
