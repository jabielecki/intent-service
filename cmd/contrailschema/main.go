package main

import (
	"github.com/tungstenfabric-preview/intent-service/pkg/cmd/contrailschema"
	log "github.com/sirupsen/logrus"
)

func main() {
	err := contrailschema.ContrailSchema.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
