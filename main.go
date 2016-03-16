package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/fheng/scm-go/config"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/fheng/scm-go/logger"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/fheng/scm-go/routes"
)

func main() {
	config := SetUpConfig()
	logger.InitLogger(config.Logger)
	SetUpDirectories(config)
	router := routes.SetUpRoutes(logger.Logger, config)
	http.Handle("/", routes.RouterPanicHandler(router, logger.Logger))
	negroni := routes.AllRoutesMiddleware(logger.Logger)
	negroni.UseHandler(router)
	listenOn := fmt.Sprintf(":%d", config.Githubtrigger.Port)
	negroni.Run(listenOn)
}

func SetUpConfig() *config.Config {
	cLineArgs := os.Args[1:]
	var confFile string
	if len(cLineArgs) > 0 {
		confFile = cLineArgs[0]
	} else {
		confFile = "./config/dev.json"
	}
	fmt.Println("config file set to ", confFile)
	return config.NewConfig(confFile)
}

func SetUpDirectories(conf *config.Config) {

	if _, err := os.Stat(conf.GetKeysPath().URL); err != nil {
		if err = os.MkdirAll(conf.GetKeysPath().URL, 0755); err != nil {
			log.Fatalln("failed to set up dirs " + err.Error())
		}
	}

	if _, err := os.Stat(conf.GetArchivePath()); err != nil {
		if err = os.MkdirAll(conf.GetArchivePath(), 0755); err != nil {
			log.Fatalln("failed to set up dirs " + err.Error())
		}
	}
}
