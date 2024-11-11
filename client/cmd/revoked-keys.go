package cmd

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/sysarmor/guard/server/pkg/apis"
)

type revokedKeys struct {
	guard

	revokeKeys string
}

func newRevokedKeys(config *Config, guard apis.Guard) *cobra.Command {
	revokedKeys := &revokedKeys{}

	command := &cobra.Command{
		Use:   "revoke-keys",
		Short: "Update the SSH certificate revocation list (CRL)",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			revokedKeys.revokeKeys = config.RevokeKeys

			if revokedKeys.revokeKeys == "" {
				return fmt.Errorf("revoked keys file is empty")
			}

			revokedKeys.guard.Guard = guard
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			err := revokedKeys.run(cmd.Context())
			if err != nil {
				slog.Error("Failed to update revoked keys",
					"error", err,
				)
			}

			return err
		},
	}

	return command
}

func (r *revokedKeys) run(ctx context.Context) error {
	slog.Info("Start to update revoked keys")
	defer slog.Info("Finish update revoked keys")

	remoteRevokedKeys, err := r.getRemoteRevokedKeys(ctx)
	if err != nil {
		return fmt.Errorf("failed to get remote revoked keys: %w", err)
	}

	fd, err := os.OpenFile(r.revokeKeys, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open revoked keys file: %w", err)
	}
	defer fd.Close()

	_, err = fd.Write(remoteRevokedKeys)
	if err != nil {
		return fmt.Errorf("failed to write revoked keys: %w", err)
	}

	return nil
}

func (r *revokedKeys) getRemoteRevokedKeys(ctx context.Context) ([]byte, error) {
	revokedKeys, err := r.GetKRL(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get remote revoked keys: %w", err)
	}

	if revokedKeys == "" {
		return nil, nil
	}

	return base64.StdEncoding.DecodeString(revokedKeys)
}
