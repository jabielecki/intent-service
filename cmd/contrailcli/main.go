package main

import (
	"github.com/tungstenfabric-preview/intent-service/pkg/cmd/contrailcli"
	log "github.com/sirupsen/logrus"
)

func main() {
	err := contrailcli.ContrailCLI.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
