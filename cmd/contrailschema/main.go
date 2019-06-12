package main

import (
	"github.com/tungstenfabric-preview/intent-service/pkg/cmd/contrailschema"
	"github.com/tungstenfabric-preview/intent-service/pkg/logutil"
)

func main() {
	err := contrailschema.ContrailSchema.Execute()
	if err != nil {
		logutil.FatalWithStackTrace(err)
	}
}
