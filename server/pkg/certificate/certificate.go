package certificate

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sysarmor/guard/server/pkg/helper"
	"golang.org/x/crypto/ssh"
)

type Certificate struct {
	privateKey []byte
	publicKey  []byte
}

func New(privateKey, publicKey []byte) *Certificate {
	return &Certificate{
		privateKey: privateKey,
		publicKey:  publicKey,
	}
}

var extensions = []string{
	"permit-X11-forwarding",
	"permit-agent-forwarding",
	"permit-port-forwarding",
	"permit-pty",
	"permit-user-rc",
}

// SignCert signs a certificate
func (c *Certificate) SignCert(
	passphrase []byte,
	publicKey []byte,
	serial uint64, id, principal string,
	validAfter uint64, validBefore uint64,
) ([]byte, error) {
	pubKey, _, _, _, err := ssh.ParseAuthorizedKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	caSigner, err := ssh.ParsePrivateKeyWithPassphrase(c.privateKey, passphrase)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	cert := &ssh.Certificate{
		Key:             pubKey,
		Serial:          serial,
		CertType:        ssh.UserCert,
		KeyId:           id,
		ValidPrincipals: []string{principal},
		ValidAfter:      validAfter,
		ValidBefore:     validBefore,
		Permissions: ssh.Permissions{
			Extensions: map[string]string{},
		},
	}

	for _, ext := range extensions {
		cert.Permissions.Extensions[ext] = ""
	}

	if err := cert.SignCert(rand.Reader, caSigner); err != nil {
		return nil, fmt.Errorf("failed to sign cert: %w", err)
	}

	return ssh.MarshalAuthorizedKey(cert), nil
}

const (
	revokedKeysFile = "revoked-keys"
	caKeyFile       = "ca.pub"
	revokeListFile  = "list-to-revoke"
)

// RevokeKeys revokes a list of keys
// golang.org/x/crypto/ssh not support revoke keys, use ssh-keygen -k to revoke keys
// e.g ssh-keygen -k -f revoked-keys -s ca list-to-revoke
// FIXME: if golang.org/x/crypto/ssh support revoke keys, replace this function
func (c *Certificate) RevokeKeys(serialIDs ...int64) ([]byte, error) {
	tempDir := os.TempDir()
	tempDir = filepath.Join(tempDir, "guard_"+helper.RandString(4))
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	revokeList := filepath.Join(tempDir, revokeListFile)
	caKeyFile := filepath.Join(tempDir, caKeyFile)
	revokedKeys := filepath.Join(tempDir, revokedKeysFile)

	revokeListFd, err := os.Create(revokeList)
	if err != nil {
		return nil, fmt.Errorf("failed to create revoked keys: %w", err)
	}
	defer revokeListFd.Close()

	for _, serialID := range serialIDs {
		if _, err := revokeListFd.WriteString(fmt.Sprintf("serial: %d\n", serialID)); err != nil {
			return nil, fmt.Errorf("failed to write revoked keys: %w", err)
		}
	}

	// write ca private key
	caFd, err := os.Create(caKeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create ca private key: %w", err)
	}
	defer caFd.Close()

	if _, err := caFd.Write(c.publicKey); err != nil {
		return nil, fmt.Errorf("failed to write ca private key: %w", err)
	}

	buf := &bytes.Buffer{}

	cmdPath, err := witchSSHKeyGen()
	if err != nil {
		return nil, fmt.Errorf("failed to find ssh-keygen: %w", err)
	}

	// revoke keys
	cmd := exec.Cmd{
		Dir:    tempDir,
		Path:   cmdPath,
		Args:   []string{"", "-k", "-f", revokedKeysFile, "-s", caKeyFile, revokeListFile},
		Stdout: buf,
		Stderr: buf,
	}

	slog.Debug("revoke keys", "cmd", cmd.String())

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to revoke keys: %w, output: %s", err, buf.String())
	}

	revokedKeysBytes, err := os.ReadFile(revokedKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to read revoked keys: %w", err)
	}

	return revokedKeysBytes, nil
}

func witchSSHKeyGen() (string, error) {
	path, err := exec.LookPath("ssh-keygen")
	if err != nil {
		return "", fmt.Errorf("failed to find ssh-keygen: %w", err)
	}
	return path, nil
}
