package cmd

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// Manage the sshd config, including the authorized principals
// file path and the trusted user CA keys file path.

type sshdConfig struct {
	sshdConfigDir string
	fileName      string

	authPrincipalsFile string
	caPubFile          string
	revokedKeys        string
}

func newSSHDConfig(sshdConfigDir, fileName *string) *cobra.Command {
	sshdConfig := &sshdConfig{}

	command := &cobra.Command{
		Use:   "init-sshd-config",
		Short: "Initialize sshd config",

		RunE: func(cmd *cobra.Command, args []string) error {
			sshdConfig.sshdConfigDir = *sshdConfigDir
			sshdConfig.fileName = *fileName

			err := sshdConfig.run()
			if err != nil {
				slog.Error("Failed to run sshd config",
					"error", err,
				)
			}

			return err
		},
	}

	flags := command.PersistentFlags()
	flags.StringVarP(&sshdConfig.authPrincipalsFile, "auth-principals-file", "", "/etc/ssh/auth_principals/%u", "The authorized principals file, default is /etc/ssh/auth_principals/%u")
	flags.StringVarP(&sshdConfig.caPubFile, "ca-pub-file", "", "/etc/ssh/guard.pub", "The trusted user CA keys file, default is /etc/ssh/guard.pub")
	flags.StringVarP(&sshdConfig.revokedKeys, "revoked-keys", "", "/etc/ssh/sshd_config.d/revoked-keys", "The revoked keys file, default is /etc/ssh/sshd_config.d/revoked-keys")

	return command
}

func (r *sshdConfig) run() error {
	err := r.initSSHConfig()
	if err != nil {
		return fmt.Errorf("failed to init ssh config: %w", err)
	}

	return nil
}

func (r *sshdConfig) initSSHConfig() error {
	path := filepath.Join(r.sshdConfigDir, r.fileName)

	// Create the directory if it does not exist
	fd, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, fs.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", path, err)
	}
	defer fd.Close()

	config := Config{}
	config.AuthorizedPrincipalsFile = r.authPrincipalsFile
	config.TrustedUserCAKeys = r.caPubFile
	config.RevokeKeys = r.revokedKeys

	err = config.WriteToFile(fd)
	if err != nil {
		return fmt.Errorf("failed to write ssh config to file %s: %w", path, err)
	}

	slog.Info("Successfully wrote ssh config to file",
		"file", path,
		"AuthorizedPrincipalsFile", config.AuthorizedPrincipalsFile,
		"TrustedUserCAKeys", config.TrustedUserCAKeys,
		"RevokedKeys", r.revokedKeys,
	)
	return nil
}

// ----  SSH Config ----
type Config struct {
	TrustedUserCAKeys        string
	AuthorizedPrincipalsFile string
	RevokeKeys               string
}

func NewConfig(trustedUserCAKeys, authorizedPrincipalsFile, revokeKeys string) *Config {
	return &Config{
		TrustedUserCAKeys:        trustedUserCAKeys,
		AuthorizedPrincipalsFile: authorizedPrincipalsFile,
		RevokeKeys:               revokeKeys,
	}
}

// ReadFromFile reads the configuration from the given reader.
func (c *Config) ReadFromFile(r io.Reader) error {
	buf := bytes.NewBuffer(nil)
	_, err := buf.ReadFrom(r)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse the config file
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// make sure the last line is parsed
				if err := c.parseLine(line); err != nil {
					return fmt.Errorf("failed to parse line: %w", err)
				}
				break
			}
			return fmt.Errorf("failed to read line from config file: %w", err)
		}

		// Skip comments
		if line[0] == '#' {
			continue
		}

		// Parse the line
		if err := c.parseLine(line); err != nil {
			return fmt.Errorf("failed to parse line: %w", err)
		}
	}

	return nil
}

// parseLine parses a single line from the config file.
func (c *Config) parseLine(line string) error {
	line = strings.TrimSuffix(line, "\n")
	if line == "" {
		return nil
	}

	parts := strings.Split(line, " ")
	if len(parts) < 2 {
		return fmt.Errorf("invalid line: %s", line)
	}

	switch parts[0] {
	case "TrustedUserCAKeys":
		c.TrustedUserCAKeys = parts[1]
	case "AuthorizedPrincipalsFile":
		c.AuthorizedPrincipalsFile = parts[1]
	case "RevokedKeys":
		c.RevokeKeys = parts[1]
	default:
		return fmt.Errorf("unknown config option: %s", parts[0])
	}

	return nil
}

var defaultComment = `# This file is used to configure the SSH server.
# It should not be edited manually.
# For more information, see sshd_config(5).
`

// WriteToFile writes the configuration to the given writer.
func (c *Config) WriteToFile(w io.Writer) error {
	_, err := fmt.Fprint(w, defaultComment)
	if err != nil {
		return fmt.Errorf("failed to write default comment: %w", err)
	}

	_, err = fmt.Fprintf(w, "TrustedUserCAKeys %s\n", c.TrustedUserCAKeys)
	if err != nil {
		return fmt.Errorf("failed to write TrustedUserCAKeys: %w", err)
	}

	_, err = fmt.Fprintf(w, "AuthorizedPrincipalsFile %s\n", c.AuthorizedPrincipalsFile)
	if err != nil {
		return fmt.Errorf("failed to write AuthorizedPrincipalsFile: %w", err)
	}

	_, err = fmt.Fprintf(w, "RevokedKeys %s\n", c.RevokeKeys)
	if err != nil {
		return fmt.Errorf("failed to write RevokeKeys: %w", err)
	}

	return nil
}
