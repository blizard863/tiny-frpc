package gssh

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

func TestGenerateED25519Key_Success(t *testing.T) {
	require := require.New(t)

	tempDir := t.TempDir()
	privPath := filepath.Join(tempDir, "id_ed25519")
	pubPath := privPath + ".pub"

	err := generateED25519Key(privPath)
	require.NoError(err)

	// 私钥文件权限
	info, err := os.Stat(privPath)
	require.NoError(err)
	if runtime.GOOS != "windows" {
		require.Equal(os.FileMode(0600), info.Mode().Perm())
	}

	// 公钥文件权限
	pubInfo, err := os.Stat(pubPath)
	require.NoError(err)
	if runtime.GOOS != "windows" {
		require.Equal(os.FileMode(0644), pubInfo.Mode().Perm())
	}

	// 私钥可正常解析（无密码）
	privBytes, err := os.ReadFile(privPath)
	require.NoError(err)
	signer, err := ssh.ParsePrivateKey(privBytes)
	require.NoError(err)
	require.NotNil(signer)

	// 公钥格式正确且与私钥匹配
	pubBytes, err := os.ReadFile(pubPath)
	require.NoError(err)
	require.True(strings.HasPrefix(string(pubBytes), "ssh-ed25519 "))

	pubKey, _, _, _, err := ssh.ParseAuthorizedKey(pubBytes)
	require.NoError(err)
	require.Equal(pubKey.Marshal(), signer.PublicKey().Marshal())
}

func TestGenerateED25519Key_InvalidPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skip permission test on windows")
	}
	require := require.New(t)

	tempDir := t.TempDir()
	require.NoError(os.Chmod(tempDir, 0444))
	defer os.Chmod(tempDir, 0755)

	err := generateED25519Key(filepath.Join(tempDir, "key"))
	require.Error(err)
	require.Contains(err.Error(), "failed to write private key file")
}

func TestGenerateED25519Key_WorksWithPublicKeyAuthFunc(t *testing.T) {
	require := require.New(t)

	tempDir := t.TempDir()
	privPath := filepath.Join(tempDir, "id_ed25519")

	require.NoError(generateED25519Key(privPath))

	auth, err := publicKeyAuthFunc(privPath)
	require.NoError(err)
	require.NotNil(auth)
}

func TestGetDefaultPrivateKeyPath_PrefersED25519(t *testing.T) {
	require := require.New(t)

	tempDir := t.TempDir()
	sshDir := filepath.Join(tempDir, ".ssh")
	require.NoError(os.MkdirAll(sshDir, 0700))

	// 同时存在两个密钥，应返回 ed25519
	require.NoError(os.WriteFile(filepath.Join(sshDir, "id_rsa"), []byte("rsa"), 0600))
	require.NoError(os.WriteFile(filepath.Join(sshDir, "id_ed25519"), []byte("ed"), 0600))

	path, err := findExistingKey(sshDir)
	require.NoError(err)
	require.Equal(filepath.Join(sshDir, "id_ed25519"), path)
}

func TestGetDefaultPrivateKeyPath_FallbackToRSA(t *testing.T) {
	require := require.New(t)

	tempDir := t.TempDir()
	sshDir := filepath.Join(tempDir, ".ssh")
	require.NoError(os.MkdirAll(sshDir, 0700))

	// 只有 rsa，应返回 rsa
	require.NoError(os.WriteFile(filepath.Join(sshDir, "id_rsa"), []byte("rsa"), 0600))

	path, err := findExistingKey(sshDir)
	require.NoError(err)
	require.Equal(filepath.Join(sshDir, "id_rsa"), path)
}

func TestGetDefaultPrivateKeyPath_NoneExists(t *testing.T) {
	require := require.New(t)

	tempDir := t.TempDir()
	sshDir := filepath.Join(tempDir, ".ssh")
	require.NoError(os.MkdirAll(sshDir, 0700))

	path, err := findExistingKey(sshDir)
	require.NoError(err)
	require.Empty(path)
}
