package config

import (
	"encoding/json"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/fheng/scm-go/data"
	"io/ioutil"
	"log"
	"os"
)

//This marshals the json config file into a useful type structure to use throughout the application

type Logger struct {
	Level    string `json:"level"`
	Filename string `json:"filename"`
}

type Connect struct {
	Level string `json:"level"`
}

type SSL struct {
	UseSSL     bool   `json:"use_ssl"`
	PrivateKey string `json:"privateKey"`
	PublicCert string `json:"publicCert"`
}

type FileServer struct {
	Path   string `json:"path"`
	Backup string `json:"backup"`
}

type Millicore struct {
	Api_bump_version string `json:"api_bump_version"`
	Url              string `json:"url"`
	Enabled          bool   `json:"enabled"`
}

type GitHubTrigger struct {
	Port    int     `json:"port"`
	Connect Connect `json:"connect"`
}

type Config struct {
	Githubtrigger GitHubTrigger `json:"githubtrigger"`
	Ssl           SSL           `json:"ssl"`
	Fileserver    FileServer    `json:"fileserver"`
	Millicore     Millicore     `json:"millicore"`
	Logger        []Logger      `json:"logger"`
}

func (c Config) GetRepoPath(repoName string) string {
	return c.Fileserver.Path + repoName
}

func (c Config) GetArchivePath() string {
	return c.Fileserver.Path + "archives"
}

func (c Config) GetKeysPath() data.KeyStore {
	return data.KeyStore{
		StoreType: data.STORE_TYPE_DISK,
		URL:       c.Fileserver.Path + "keys/",
	}
}

var Conf *Config

//constructor for getting a new config . Reads the passed file path and converts it from json into a *Config
func NewConfig(filePath string) *Config {

	var (
		congFile *os.File
		err      error
		confData []byte
		conf     *Config
	)

	if congFile, err = os.Open(filePath); err != nil {
		log.Panic("failed to open config file " + err.Error())
	}

	if confData, err = ioutil.ReadAll(congFile); err != nil {
		log.Panic("failed to read config file " + err.Error())
	}

	if err = json.Unmarshal(confData, &conf); err != nil {
		log.Panic("failed to parse json config data " + err.Error())
	}
	Conf = conf
	return conf
}

func GetConfig() *Config {
	if nil == Conf {
		log.Fatal("config has not been initialised. ")
	}
	return Conf
}
