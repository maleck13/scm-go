package config_test

import (
	"encoding/json"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/fheng/scm-go/config"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func Test_Marshal_Config(t *testing.T) {
	conf := `{
		"githubtrigger": {
			"port": 8801,
			"connect": {
				"level": "log4js.levels.INFO"
			}
		},
		"ssl": {
			"use_ssl": false,
			"privateKey":"",
			"publicCert":""
		},
		"fileserver": {
    		"path": "/tmp/fh-scm",
    		"backup": "/tmp/fh-scm/backup"
  		},
  		"millicore": {
    		"api_bump_version": "https://CLUSTER_HOST/box/srv/1.1/pub/app/APP_GUID/bump",
    		"url": "https://localhost"
  		}
	}`

	var confCont config.Config

	if err := json.Unmarshal([]byte(conf), &confCont); err != nil {
		log.Fatal(err.Error())
	}

	log.Println(confCont)

	log.Println(confCont)

}

func Test_NewConfig(t *testing.T) {
	conf := config.NewConfig("../config/dev.json")
	if nil == conf {
		t.Fail()
	}
	assert.Equal(t, "/tmp/fh-scm/test", conf.GetRepoPath("test"), "expected repo name to be correct")
	assert.Equal(t, 8801, conf.Githubtrigger.Port, " port should be 8801")
	assert.Equal(t, "debug", conf.Logger[0].Level, " level should debug")

}
