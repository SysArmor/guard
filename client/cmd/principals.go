package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/sysarmor/guard/server/pkg/apis"
	"github.com/sysarmor/guard/server/pkg/apis/dto"
)

type principals struct {
	guard

	authorizedPrincipalsFile string
}

func newPrincipals(config *Config, guard apis.Guard) *cobra.Command {
	principals := &principals{}

	command := &cobra.Command{
		Use:   "principals",
		Short: "Update the authorized principals for SSH certificates",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			principals.authorizedPrincipalsFile = strings.TrimSuffix(config.AuthorizedPrincipalsFile, "%u")

			if principals.authorizedPrincipalsFile == "" {
				return fmt.Errorf("authorized principals file is empty")
			}

			principals.guard.Guard = guard
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			err := principals.run(cmd.Context())
			if err != nil {
				slog.Error("Failed to update principals",
					"error", err,
				)
			}

			return err
		},
	}

	return command
}

func (r *principals) run(ctx context.Context) error {
	slog.Info("Start to update principals")
	defer slog.Info("Finish update principals")

	remotePrincipals, err := r.getRemotePrincipals(ctx)
	if err != nil {
		return fmt.Errorf("failed to get remote principals: %w", err)
	}

	localPrincipals, err := r.getLocalPrincipals()
	if err != nil {
		return fmt.Errorf("failed to get local principals: %w", err)
	}

	// prepare the authorized principals directory
	err = os.MkdirAll(r.authorizedPrincipalsFile, 0755)
	if err != nil {
		return fmt.Errorf("failed to create authorized principals directory: %w", err)
	}

	// compare the remote and local principals
	// if different, update the local principals
	// if the role is not exist in local, create it
	for _, remote := range remotePrincipals {
		var needUpdate bool
		local, ok := r.findLocalPrincipals(localPrincipals, remote.Role)
		if ok {
			diff, ok := r.comparePrincipals(local.Principals, remote.Principals)
			if !ok {
				needUpdate = !ok
				slog.Info("principals is different",
					"role", remote.Role,
					"diff", diff,
				)
			}
		}

		if !ok || needUpdate {
			err := r.updateLocalPrincipals(remote.Role, remote.Principals)
			if err != nil {
				return fmt.Errorf("failed to update local principals: %w", err)
			}

			slog.Info("principals updated",
				"role", remote.Role,
			)
			continue
		}
	}

	return nil
}

// findLocalPrincipals find the local principals by role
func (r *principals) findLocalPrincipals(principals []dto.Principals, role string) (dto.Principals, bool) {
	for _, principal := range principals {
		if principal.Role == role {
			return principal, true
		}
	}

	return dto.Principals{}, false
}

// comparePrincipals compare the local and remote principals
func (r *principals) comparePrincipals(local, remote []string) ([]string, bool) {

	diff := make([]string, 0)

	for _, val := range remote {
		if !slices.Contains(local, val) {
			diff = append(diff, "+"+val)
		}
	}

	for _, val := range local {
		if !slices.Contains(remote, val) {
			diff = append(diff, "-"+val)
		}
	}

	return diff, len(diff) == 0
}

// getRemotePrincipals get the principals from server
func (r *principals) getRemotePrincipals(ctx context.Context) ([]*dto.Principals, error) {
	principals, err := r.GetPrincipals(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get principals from server: %w", err)
	}

	return principals, nil
}

// getLocalPrincipals get the principals from local
func (r *principals) getLocalPrincipals() ([]dto.Principals, error) {
	principals := make([]dto.Principals, 0)

	err := filepath.WalkDir(r.authorizedPrincipalsFile, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("failed to walk dir: %w", err)
		}

		if d.IsDir() {
			// do not support deep directory
			return nil
		}

		body, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		principals = append(principals, dto.Principals{
			Role:       d.Name(),
			Principals: r.parsePrincipals(string(body)),
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get local principals: %w", err)
	}

	return principals, nil
}

// parsePrincipals parse the body of authorized principals file
// and return the principals
func (r *principals) parsePrincipals(body string) []string {
	principals := make([]string, 0)

	for _, val := range strings.Split(body, "\n") {
		if val == "" {
			continue
		}

		// skip comment
		if strings.HasPrefix(val, "#") {
			continue
		}

		principals = append(principals, val)
	}

	return principals
}

// updateLocalPrincipals update the authorized principals file
func (r *principals) updateLocalPrincipals(role string, principals []string) error {
	file := filepath.Join(r.authorizedPrincipalsFile, role)
	fd, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer fd.Close()

	_, err = fd.WriteString(fmt.Sprintf(defaultPrincipalsComment, role))
	if err != nil {
		return fmt.Errorf("failed to write comment: %w", err)
	}
	_, err = fd.WriteString(r.serializePrincipals(principals))
	if err != nil {
		return fmt.Errorf("failed to write principals: %w", err)
	}

	return nil
}

var defaultPrincipalsComment = `# Authorized principals for role %s
# This file is managed by guard, do not edit it manually
`

func (r *principals) serializePrincipals(principals []string) string {
	return strings.Join(principals, "\n")
}
