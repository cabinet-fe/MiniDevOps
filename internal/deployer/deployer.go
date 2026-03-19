package deployer

import "context"

type ServerInfo struct {
	Host       string
	Port       int
	Username   string
	AuthType   string
	Password   string
	PrivateKey string
}

type DeployOptions struct {
	SourceDir  string
	Server     ServerInfo
	RemotePath string
	Logger     func(string)
}

type Deployer interface {
	Deploy(ctx context.Context, opts DeployOptions) error
}

func NewDeployer(method string) Deployer {
	switch method {
	case "rsync":
		return &RsyncDeployer{}
	case "sftp":
		return &SFTPDeployer{}
	case "scp":
		return &SCPDeployer{}
	default:
		return &RsyncDeployer{}
	}
}
