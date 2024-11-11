package cmd

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/sysarmor/guard/server/pkg/apis"
)

// manage the ca File
type ca struct {
	trustedUserCAKeys string

	guard
}

func newCA(config *Config, guard apis.Guard) *cobra.Command {
	ca := &ca{}

	command := &cobra.Command{
		Use:   "ca",
		Short: "Update the trusted user CA keys file (guard.pub)",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			ca.trustedUserCAKeys = config.TrustedUserCAKeys
			if ca.trustedUserCAKeys == "" {
				return fmt.Errorf("trusted user ca keys is empty")
			}

			ca.guard.Guard = guard
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			err := ca.run(cmd.Context())
			if err != nil {
				slog.Error("Failed to run ca",
					"error", err,
				)
			}
			return err
		},
	}

	return command
}

func (ca *ca) run(ctx context.Context) error {
	remoteCAPub, err := ca.getCAFromServer(ctx)
	if err != nil {
		return fmt.Errorf("failed to get guard.pub from server: %w", err)
	}

	localCAPub, err := ca.getCAFromLocal()
	if err != nil {
		return fmt.Errorf("failed to get guard.pub from local: %w", err)
	}

	if remoteCAPub != localCAPub {
		err := ca.updateCA(remoteCAPub)
		if err != nil {
			return fmt.Errorf("failed to update guard.pub: %w", err)
		}

		slog.Info("guard.pub updated")
		return nil
	}

	slog.Info("guard.pub is up to date, no need to update")
	return nil
}

func (ca *ca) getCAFromServer(ctx context.Context) (string, error) {
	caPub, err := ca.guard.GetCA(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get guard.pub: %w", err)
	}
	return caPub, nil
}

func (ca *ca) getCAFromLocal() (string, error) {
	fd, err := os.OpenFile(ca.trustedUserCAKeys, os.O_RDONLY, 0644)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("failed to open guard.pub file: %w", err)
	}
	defer fd.Close()

	buf := bytes.NewBuffer(nil)
	_, err = buf.ReadFrom(fd)
	if err != nil {
		return "", fmt.Errorf("failed to read guard.pub file: %w", err)
	}

	return buf.String(), nil
}

func (ca *ca) updateCA(caPub string) error {
	fd, err := os.OpenFile(ca.trustedUserCAKeys, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open guard.pub file: %w", err)
	}
	defer fd.Close()

	_, err = fd.WriteString(caPub)
	if err != nil {
		return fmt.Errorf("failed to write guard.pub file: %w", err)
	}

	return nil
}
