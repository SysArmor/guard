package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/robfig/cron"
	"github.com/spf13/cobra"
	"github.com/sysarmor/guard/server/pkg/apis"
)

type daemon struct {
	config *Config

	section []string
	cron    string
}

func newDaemon(config *Config, guard apis.Guard) *cobra.Command {
	daemon := &daemon{
		config: config,
	}

	command := &cobra.Command{
		Use:   "daemon",
		Short: "Start the guard daemon to manage SSH CA, principals, and revoked keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := daemon.daemonize(); err != nil {
				return fmt.Errorf("failed to daemonize: %w", err)
			}
			return daemon.run(cmd.Context(), guard)
		},
	}

	command.Flags().StringArrayVarP(&daemon.section, "section", "s", []string{"all"}, "The section to run, default is all, available sections: all, ca, principals, revoke-keys")
	command.Flags().StringVarP(&daemon.cron, "cron", "c", "0 0/5 * * *", "The cron expression to run the daemon, default is every 5 minutes")

	return command
}

// daemonize daemonize the process
func (d *daemon) daemonize() error {
	if os.Getppid() != 1 {
		cmd := exec.Command(os.Args[0], os.Args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setsid: true,
		}
		slog.Info("Daemonize the process")
		if err := cmd.Start(); err != nil {
			return err
		}
		os.Exit(0)
	}
	return nil
}

// fn is a function type
type fn func(ctx context.Context) error

// run runs the daemon
func (d *daemon) run(ctx context.Context, guard apis.Guard) error {
	fns, err := d.handleSection(ctx, guard)
	if err != nil {
		return fmt.Errorf("failed to handle section: %w", err)
	}

	cron := cron.New()
	defer cron.Stop()

	fn := func(ctx context.Context) {
		for _, fn := range fns {
			err := fn(ctx)
			if err != nil {
				slog.Error("Failed to run function",
					"error", err,
				)
			}
		}
	}

	fn(ctx)
	cron.AddFunc(d.cron, func() {
		fn(ctx)
	})
	cron.Start()

	<-ctx.Done()
	slog.Info("Daemon stopped")
	return nil
}

// handleSection handles the section flag
func (d *daemon) handleSection(ctx context.Context, guard apis.Guard) ([]fn, error) {
	fns := make([]fn, 0, 2)
	for _, section := range d.section {
		switch section {
		case "all":
			d.section = []string{"ca", "principals", "revoke-keys"}
			return d.handleSection(ctx, guard)
		case "ca":
			fns = append(fns, func(ctx context.Context) error {
				ca := ca{}
				ca.guard.Guard = guard
				ca.trustedUserCAKeys = strings.TrimSuffix(d.config.TrustedUserCAKeys, "%u")

				return ca.run(ctx)
			})
		case "principals":
			fns = append(fns, func(ctx context.Context) error {
				principals := principals{}
				principals.guard.Guard = guard
				principals.authorizedPrincipalsFile = strings.TrimSuffix(d.config.AuthorizedPrincipalsFile, "%u")

				return principals.run(ctx)
			})
		case "revoke-keys":
			fns = append(fns, func(ctx context.Context) error {
				revokeKeys := revokedKeys{}
				revokeKeys.guard.Guard = guard
				revokeKeys.revokeKeys = d.config.RevokeKeys

				return revokeKeys.run(ctx)
			})
		default:
			return nil, fmt.Errorf("unsupported section: %s", d.section)
		}
	}

	return fns, nil
}
