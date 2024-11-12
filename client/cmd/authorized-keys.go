package cmd

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/sysarmor/guard/server/pkg/apis"
)

type authorizedKeys struct {
	guard

	authorizedKeysPath string
}

func newAuthorizedKeys(_ *Config, guard apis.Guard) *cobra.Command {
	publicKey := &authorizedKeys{}

	command := &cobra.Command{
		Use:   "authorized-keys",
		Short: "Update the authorized keys for SSH certificates",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			home := os.Getenv("HOME")
			if home == "" {
				return fmt.Errorf("HOME environment variable is not set")
			}
			publicKey.authorizedKeysPath = filepath.Join(home, ".ssh/guard_keys")

			publicKey.guard.Guard = guard
			return nil
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			err := publicKey.run(cmd.Context())
			if err != nil {
				slog.Error("Failed to update public key",
					"error", err,
				)
			}

			return err
		},
	}

	flags := command.PersistentFlags()
	flags.StringVarP(&publicKey.authorizedKeysPath, "authorized-keys-path", "", "", "The authorized keys path, default is $HOME/.ssh/guard_keys")

	return command
}

func (r *authorizedKeys) run(ctx context.Context) error {
	slog.Info("Updating authorized keys")
	defer slog.Info("Finished updating authorized keys")

	remoteAuthorizedKeys, err := r.getRemoteAuthorizedKeys(ctx)
	if err != nil {
		return fmt.Errorf("failed to get remote authorized keys: %w", err)
	}

	same, err := r.compareAuthorizedKeys(remoteAuthorizedKeys)
	if err != nil {
		return fmt.Errorf("failed to compare authorized keys: %w", err)
	}

	if same {
		slog.Info("Authorized keys are the same, no need to update")
	}

	return r.overwriteAuthorizedKeys(remoteAuthorizedKeys)
}

func (r *authorizedKeys) overwriteAuthorizedKeys(remote []string) error {
	fd, err := os.OpenFile(r.authorizedKeysPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open authorized keys file: %w", err)
	}
	defer fd.Close()

	for _, key := range remote {
		_, err = fd.WriteString(key + "\n")
		if err != nil {
			return fmt.Errorf("failed to write authorized key: %w", err)
		}
	}

	return nil
}

func (r *authorizedKeys) getRemoteAuthorizedKeys(ctx context.Context) ([]string, error) {
	return r.guard.Guard.GetAuthorizedKeys(ctx)
}

func (r *authorizedKeys) getLocalAuthorizedKeysMD5() (string, error) {
	fd, err := os.OpenFile(r.authorizedKeysPath, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to open file %s: %w", r.authorizedKeysPath, err)
	}
	defer fd.Close()

	buf := bytes.NewBuffer(nil)
	_, err = buf.ReadFrom(fd)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", r.authorizedKeysPath, err)
	}

	return r.md5(buf.Bytes()), nil
}

func (r *authorizedKeys) md5(body []byte) string {
	hash := md5.New()
	hash.Write(body)
	return hex.EncodeToString(hash.Sum(nil))
}

func (r *authorizedKeys) compareAuthorizedKeys(remote []string) (bool, error) {
	buf := bytes.NewBuffer(nil)
	for _, key := range remote {
		buf.WriteString(key)
		buf.WriteString("\n")
	}

	localMd5, err := r.getLocalAuthorizedKeysMD5()
	if err != nil {
		return false, fmt.Errorf("failed to get local authorized keys md5: %w", err)
	}

	remoteMd5 := r.md5(buf.Bytes())
	return localMd5 == remoteMd5, nil
}
