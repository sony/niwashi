// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright 2025 Sony Group Corporation

package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/sony/niwashi/internal/action"
	"github.com/sony/niwashi/internal/file"
	"github.com/sony/niwashi/internal/state"
	"github.com/sony/niwashi/internal/transport/ssh"
	"github.com/sony/niwashi/internal/workspace"
	"github.com/spf13/cobra"
)

type SshData struct {
	WorkDir  string
	NodeName string
}

var sshData SshData

// sshCmd represents the ssh command
func newSshCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ssh NODE_NAME",
		Short: "Connect to a node via SSH",
		Long: `Connects to a managed node via SSH using credentials automatically
retrieved from the state file. It simplifies access by eliminating the
need to manually look up IP addresses or key paths.
`,
		SilenceUsage: true, // Don't show usage on error
		Args:         cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sshData.NodeName = args[0]
			return RunSshCmd(&sshData)
		},
	}

	cmd.Flags().StringVar(&sshData.WorkDir, FlagWorkDir, DefaultWorkDir, UsageWorkDir)

	return cmd
}

func RunSshCmd(data *SshData) (err error) {

	fs := file.NewDefaultFileSystem()
	ws := workspace.NewRunWs(data.WorkDir)
	reader, err := fs.Open(ws.GetStateFilePath())
	if err != nil {
		return err
	}

	s, err := state.LoadAsJson(reader)
	err = errors.Join(err, reader.Close())
	if err != nil {
		return err
	}

	connection, err := s.GetConnection(data.NodeName).GetTransport(ssh.TypeName)
	if err != nil {
		return fmt.Errorf("failed to get SSH transport for node: %s, error: %v", data.NodeName, err)
	}
	c := connection.Clone().(*ssh.Transport)
	if c == nil {
		return fmt.Errorf("no SSH transport found for node: %s", data.NodeName)
	}

	// Update known_hosts path with rendered template
	te := action.NewDefaultTemplateEngine(&action.TemplateParams{
		Paths: map[string]string{
			"workspace": data.WorkDir,
		},
	})
	knownHostsPath, err := te.Render(c.HostKey.KnownHostsPath)
	if err != nil {
		return err
	}
	c.HostKey.KnownHostsPath = knownHostsPath

	client, err := ssh.NewSshClient(context.Background(), c, fs)
	if err != nil {
		return err
	}
	defer func() {
		if e := client.Close(); e != nil {
			err = errors.Join(err, e)
		}
	}()

	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer func() {
		_ = session.Close()
	}()

	session.SetStderr(os.Stderr)
	session.SetStdout(os.Stdout)
	session.SetStdin(os.Stdin)

	if err := session.Shell(); err != nil {
		return err
	}

	return nil
}
