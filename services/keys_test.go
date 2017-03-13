package services_test

import (
	"github.com/maleck13/scm-go/config"
	"github.com/maleck13/scm-go/data"
	"github.com/maleck13/scm-go/services"
	"github.com/maleck13/scm-go/test"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func init() {
	config.NewConfig("../config/dev.json")
	test.SetUpDirectories(config.Conf)
}

func TestSetUpSSHKeysAndRemoveOK(t *testing.T) {

	gitParams := data.NewScmParams()
	gitParams.RepoKey = test.TestGitPrivKey
	gitParams.AppGuid = "testapp"
	pubPrivKeyLoc, err := services.SetUpSSHKeys(gitParams.RepoContext, gitParams.AppContext, config.Conf.GetKeysPath)
	if nil != err {
		t.Log(err.Error())
		t.Fail()
	}
	assert.NotNil(t, pubPrivKeyLoc, "did not expect nil cred")

	assert.NotNil(t, pubPrivKeyLoc.PrivateKeyPath, "expected private key path")
	assert.NotNil(t, pubPrivKeyLoc.PublicKeyPath, "expected public key path")
	assertFilesSame(t, pubPrivKeyLoc.PrivateKeyPath, test.TestGitPrivKey)
	assertFilesSame(t, pubPrivKeyLoc.PublicKeyPath, test.TestPubKey)

	err = services.RemoveKeys(pubPrivKeyLoc)
	assert.NoError(t, err, "did not expect an error removing keys")
	_, err = os.Stat(pubPrivKeyLoc.PrivateKeyPath)
	assert.Error(t, err, "expected an error as file should not exist")

}

func TestSetUpSSHKeysFail(t *testing.T) {
	gitParams := data.NewScmParams()
	gitParams.AppGuid = "testapp"
	_, err := services.SetUpSSHKeys(gitParams.RepoContext, gitParams.AppContext, config.Conf.GetKeysPath)
	assert.Error(t, err, "expected an error setting up keys")
}

func TestGetPublicPrivateKeyPairFromPrivateKey(t *testing.T) {

	pKeyPair, err := services.GetPublicPrivateKeyPairFromPrivateKey(test.TestGitPrivKey)
	assert.NoError(t, err, "did not expect an error")
	assert.NotNil(t, pKeyPair, "expected a publicPrivateKeyPair")
	key, err := pKeyPair.PublicKeyToMem()
	assert.NoError(t, err, "did not expect an error")
	assert.NotNil(t, key, "expected public key not to be nil")
	keyStr := strings.TrimSpace(string(key))
	assert.Equal(t, test.TestPubKey, keyStr, "expected public key to equal")

}

func TestGetPublicPrivateKeyPairFromPrivateKeyFail(t *testing.T) {

	notAKey := "notaprivatekey"
	_, err := services.GetPublicPrivateKeyPairFromPrivateKey(notAKey)
	assert.Error(t, err, "expected an error")
}

func assertFilesSame(t *testing.T, filePath, contents string) {

	file, err := os.Open(filePath)
	assert.NoError(t, err, " did not expect an error")
	content, err := ioutil.ReadAll(file)
	assert.NoError(t, err, " did not expect an error")
	key := string(content)
	assert.Equal(t, strings.TrimSpace(contents), strings.TrimSpace(key), "expected key to be the same")
}
