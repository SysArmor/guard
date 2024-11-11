package certificate

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/pem"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

// generateKeyPair generates a RSA key pair with the specified bits and comment.
func generateKeyPair(bits int, comment string, passphrase []byte) (privateKey []byte, publicKey []byte, err error) {
	priv, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}

	// generate private key
	privPEM, err := ssh.MarshalPrivateKeyWithPassphrase(priv, comment, passphrase)
	if err != nil {
		return nil, nil, err
	}

	// generate public key
	pub, err := ssh.NewPublicKey(&priv.PublicKey)
	if err != nil {
		return nil, nil, err
	}
	pubBytes := ssh.MarshalAuthorizedKey(pub)

	return pem.EncodeToMemory(privPEM), pubBytes, nil
}

func TestSignCert(t *testing.T) {
	t.Run("SignCert", func(t *testing.T) {
		var passphrase = []byte("123456")

		// generate CA
		caPrivateKey, _, err := generateKeyPair(2048, "", passphrase)
		if err != nil {
			t.Fatalf("failed to generate CA key pair: %v", err)
		}

		// generate user key pair
		_, userPublicKey, err := generateKeyPair(2048, "", passphrase)
		if err != nil {
			t.Fatalf("failed to generate user key pair: %v", err)
		}

		// sign user certificate
		serial := uint64(1)
		keyId := "went"
		principal := "admin"
		validAfter := time.Now()
		validBefore := validAfter.Add(16 * 7 * 24 * time.Hour)

		caCert := New(caPrivateKey, nil)
		signedCert, err := caCert.SignCert(passphrase, userPublicKey, serial, keyId, principal,
			uint64(validAfter.Unix()), uint64(validBefore.Unix()))
		if err != nil {
			t.Fatalf("failed to sign user certificate: %v", err)
		}

		// verify signed certificate
		_, _, _, _, err = ssh.ParseAuthorizedKey(signedCert)
		if err != nil {
			t.Fatalf("failed to parse signed certificate: %v", err)
		}

		t.Logf("Successfully signed certificate:\n%s", signedCert)
	})

	t.Run("SignWithOpenSSH", func(t *testing.T) {
		var passphrase = []byte("123456")

		// prepare private key
		// ssh-keygen -C CA -f ca -b 4096
		var caPrivateKey = []byte(`-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAACmFlczI1Ni1jdHIAAAAGYmNyeXB0AAAAGAAAABAhu1MZ3c
UetiU+SncYkWjkAAAAGAAAAAEAAAAzAAAAC3NzaC1lZDI1NTE5AAAAIAiE9ix3XLKG6lWA
NiFGTOIiQYKUzTA9lQ9Hmlj+FHX6AAAAkL/WSnxJe58oEES3/Tzzefrf1cEHcIyVfDUBjH
kVrFT4Ar6CjnkBWNyB6FgSnRWA8Hus1DNuxqPBpI3ZJ6khlUhm9aJcdVQRlh9diggjMsxJ
8p1Ym8aLX9H2ds2JoD0vh3sHV1Px0NuMdDM4Zq24J3UZOH7/YGn6ryWuZigAx60jrwwpAD
ANtqQ77jfdi/FL3Q==
-----END OPENSSH PRIVATE KEY-----`)

		// prepare public key
		// ssh-keygen -t rsa -C ""
		var publicKey = []byte(`ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCkNobhY0nOw+znjxBJhyUqDWfZNRpqz6iWSqBwoxMmTtLXtlQWG6g9/eQGD/nWpRGtoaA1pJS8qoaMzwl4CLEmOxQeSa52mnvzHsifOEF/tCKXBMbilz2q+iXX0sHTlEpTE2lAJRJX/aRFuahI9j2gICRcHGD72L05qFcaxMYoqchtvZRK8+Ltt5nIuLxxmgFzSZnDbMRDfp1st5CAaGfEJ4uqVSfeIbrKJ3s1hGcVAMKdc+Qfz1aiaRC3J0gtiQJsOmfZvaVQk0m5yiiAJK3iYHiwNmc4dPlqGtL+vwCMcPtI7CgDJNGu4wuY7FaZfMuBRKaqk96DjCZvVZSlyP9YoHvPH0hwTAYcXMnZozuElIptnt31nxZcp9j2UW+21Xpv5Ze06cCkgNIRsgYeFd3Ei/Gjzl2pbFAio21KpVbmyD+rjqtXDFlkYb9RlJC4zFWmlnKCNbKV7CcX5LEVhZhl080jGFh6RnqTM3IANGka857DS6EUqCpla1l+f6CJuSk= “”`)

		// sign user certificate
		serial := uint64(1)
		keyId := "went"
		principal := "admin"
		validAfter := time.Now()
		validBefore := validAfter.Add(16 * 7 * 24 * time.Hour)

		caCert := New(caPrivateKey, nil)
		signedCert, err := caCert.SignCert(passphrase, publicKey, serial, keyId, principal,
			uint64(validAfter.Unix()), uint64(validBefore.Unix()))
		if err != nil {
			t.Fatalf("failed to sign user certificate: %v", err)
		}

		// verify signed certificate
		_, _, _, _, err = ssh.ParseAuthorizedKey(signedCert)
		if err != nil {
			t.Fatalf("failed to parse signed certificate: %v", err)
		}

		t.Logf("Successfully signed certificate:\n%s", signedCert)
	})
}

func TestRevokeKeys(t *testing.T) {
	t.Run("RevokeKeys", func(t *testing.T) {
		var passphrase = []byte("123456")

		// generate CA
		_, caPublicKey, err := generateKeyPair(2048, "", passphrase)
		if err != nil {
			t.Fatalf("failed to generate CA key pair: %v", err)
		}

		// serial number of the certificate to be revoked
		var serial int64 = 1

		caCert := New(nil, caPublicKey)

		// revoke user certificate
		_, err = caCert.RevokeKeys(serial)
		if err != nil {
			t.Fatalf("failed to revoke user certificate: %v", err)
		}
	})
}
