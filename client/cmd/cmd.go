package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/sysarmor/guard/server/pkg/apis"
)

func New() *cobra.Command {
	sshdConfig := &Config{}

	var sshdConfigDir string
	var fileName string
	var dryRun bool

	var daemon = &daemon{}
	var guard = &guard{}

	root := cobra.Command{
		Use:   "guard-client",
		Short: "A client tool for managing SSH configurations and Guard operation",

		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if daemon.isNeedDaemonize() {
				if err := daemon.daemonize(cmd.Context()); err != nil {
					return fmt.Errorf("failed to daemonize: %w", err)
				}
			}

			if dryRun {
				guard.Guard = &apis.FakeGuard{}
			}

			path := filepath.Join(sshdConfigDir, fileName)

			// Create the directory if it does not exist
			fd, err := os.OpenFile(path, os.O_RDONLY, fs.ModePerm)
			if os.IsNotExist(err) {
				return nil
			}
			if err != nil {
				return fmt.Errorf("failed to open file %s: %w", path, err)
			}
			defer fd.Close()

			err = sshdConfig.ReadFromFile(fd)
			if err != nil {
				return fmt.Errorf("failed to read config from file: %w", err)
			}

			return nil
		},
	}

	root.AddCommand(newSSHDConfig(&sshdConfigDir, &fileName))
	root.AddCommand(newCA(sshdConfig, guard))
	root.AddCommand(newPrincipals(sshdConfig, guard))
	root.AddCommand(newRevokedKeys(sshdConfig, guard))
	root.AddCommand(newAuthorizedKeys(sshdConfig, guard))

	flags := root.PersistentFlags()
	flags.StringVarP(&sshdConfigDir, "sshd-config-dir", "", "/etc/ssh/sshd_config.d/", "The directory of sshd config files, default is /etc/ssh/sshd_config.d/")
	flags.StringVarP(&fileName, "file-name", "f", "guard.conf", "The config file name, default is guard.conf")
	flags.BoolVar(&dryRun, "dry-run", false, "Dry run mode, default is false")

	daemon.PersistentFlags(flags)
	guard.PersistentFlags(flags)

	return &root
}
