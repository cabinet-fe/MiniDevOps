package deployer

import (
	"context"
	"testing"
)

// mockDeployer exercises the Deployer interface without a live remote host.
// Live rsync/sftp/scp/agent paths need SSH targets; only local is integration-tested
// in local_test.go. Remote methods are covered via NewDeployer selection + this mock.
type mockDeployer struct {
	called bool
	opts   DeployOptions
	err    error
}

func (m *mockDeployer) Deploy(ctx context.Context, opts DeployOptions) error {
	m.called = true
	m.opts = opts
	return m.err
}

func TestNewDeployer_selectsRemoteMethods(t *testing.T) {
	t.Parallel()
	cases := map[string]any{
		"rsync": &RsyncDeployer{},
		"sftp":  &SFTPDeployer{},
		"scp":   &SCPDeployer{},
		"agent": &AgentDeployer{},
		"local": &LocalDeployer{},
	}
	for method, wantType := range cases {
		d := NewDeployer(method)
		if d == nil {
			t.Fatalf("%s: nil deployer", method)
		}
		switch wantType.(type) {
		case *RsyncDeployer:
			if _, ok := d.(*RsyncDeployer); !ok {
				t.Fatalf("%s: got %T", method, d)
			}
		case *SFTPDeployer:
			if _, ok := d.(*SFTPDeployer); !ok {
				t.Fatalf("%s: got %T", method, d)
			}
		case *SCPDeployer:
			if _, ok := d.(*SCPDeployer); !ok {
				t.Fatalf("%s: got %T", method, d)
			}
		case *AgentDeployer:
			if _, ok := d.(*AgentDeployer); !ok {
				t.Fatalf("%s: got %T", method, d)
			}
		case *LocalDeployer:
			if _, ok := d.(*LocalDeployer); !ok {
				t.Fatalf("%s: got %T", method, d)
			}
		}
	}
}

func TestDeployerInterface_mockRemoteSmoke(t *testing.T) {
	t.Parallel()
	mock := &mockDeployer{}
	var d Deployer = mock
	err := d.Deploy(context.Background(), DeployOptions{
		SourceDir:  "/tmp/src",
		RemotePath: "/var/www",
		Server: ServerInfo{
			Host:     "example.invalid",
			Port:     22,
			Username: "deploy",
			AuthType: "password",
			Password: "x",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !mock.called {
		t.Fatal("expected Deploy to be called")
	}
	if mock.opts.Server.Host != "example.invalid" {
		t.Fatalf("host=%q", mock.opts.Server.Host)
	}
}
