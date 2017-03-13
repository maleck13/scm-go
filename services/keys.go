package services

import (
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"github.com/maleck13/scm-go/data"
	"github.com/maleck13/scm-go/logger"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

type PublicPrivateKeyPair struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  crypto.PublicKey
}

type PublicPrivateKeyLocation struct {
	PrivateKeyPath string
	PublicKeyPath  string
}

func (pki *PublicPrivateKeyPair) PublicKeyToMem() ([]byte, error) {
	pKey, err := ssh.NewPublicKey(pki.PublicKey)
	if nil != err {
		return nil, err
	}
	mBytes := ssh.MarshalAuthorizedKey(pKey)

	return mBytes, nil
}

func (pki *PublicPrivateKeyPair) PrivateKeyToMem() ([]byte, error) {

	mBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(pki.PrivateKey),
	})
	return mBytes, nil
}

func SetUpSSHKeys(repoContext *data.RepoContext, appContext *data.AppContext, keyStore data.KeyStoreLocator) (PublicPrivateKeyLocation, error) {

	var publicPrivateKeyLoc PublicPrivateKeyLocation // used as value rather than pointer

	if "" == repoContext.RepoKey {
		return publicPrivateKeyLoc, errors.New("no private key in params")
	}

	publicPrivateKey, err := GetPublicPrivateKeyPairFromPrivateKey(repoContext.RepoKey)
	if err != nil {
		return publicPrivateKeyLoc, err
	}

	publicPrivateKeyLoc, err = createKeyFiles(publicPrivateKey, appContext.AppGuid, keyStore())
	if nil != err {
		return publicPrivateKeyLoc, err
	}

	return publicPrivateKeyLoc, nil

}

func createKeyFiles(pubPriv *PublicPrivateKeyPair, appId string, storeLoc data.KeyStore) (PublicPrivateKeyLocation, error) {
	var (
		pubPrivKeyLoc PublicPrivateKeyLocation
		logger        = logger.Logger
	)

	if storeLoc.StoreType != data.STORE_TYPE_DISK {
		return pubPrivKeyLoc, errors.New("unsupported key store")
	}

	if _, err := os.Stat(storeLoc.URL); err != nil {
		return pubPrivKeyLoc, err
	}
	timeStamp := time.Now().Nanosecond()
	pKeyFile := storeLoc.URL + appId + strconv.Itoa(timeStamp)
	pubKeyFile := storeLoc.URL + appId + strconv.Itoa(timeStamp) + ".pub"

	pKeyBytes, err := pubPriv.PrivateKeyToMem()
	if err != nil {
		return pubPrivKeyLoc, err
	}
	publicKeyBytes, err := pubPriv.PublicKeyToMem()
	if err != nil {
		return pubPrivKeyLoc, err
	}
	if err := ioutil.WriteFile(pKeyFile, pKeyBytes, 0600); err != nil {
		return pubPrivKeyLoc, err
	}
	if err := ioutil.WriteFile(pubKeyFile, publicKeyBytes, 0600); err != nil {
		return pubPrivKeyLoc, err
	}

	logger.Info("created keys " + pubKeyFile + " private " + pKeyFile)
	return PublicPrivateKeyLocation{PublicKeyPath: pubKeyFile, PrivateKeyPath: pKeyFile}, nil

}

func GetPublicPrivateKeyPairFromPrivateKey(key string) (*PublicPrivateKeyPair, error) {

	//note does not like if no \n after comment
	block, _ := pem.Decode([]byte(key))
	if nil == block {
		return nil, errors.New("could not decode pem file")
	}
	rsaPrivate, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return &PublicPrivateKeyPair{PublicKey: rsaPrivate.Public(), PrivateKey: rsaPrivate}, nil
}

func RemoveKeys(credLoc PublicPrivateKeyLocation) error {
	if err := os.Remove(credLoc.PrivateKeyPath); err != nil {
		return err
	}

	if err := os.Remove(credLoc.PublicKeyPath); err != nil {
		return err
	}

	return nil
}
