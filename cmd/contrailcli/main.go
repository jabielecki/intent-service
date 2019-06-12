package main

import (
	"github.com/tungstenfabric-preview/intent-service/pkg/cmd/contrailcli"
	"github.com/tungstenfabric-preview/intent-service/pkg/logutil"
)

func main() {
	err := contrailcli.ContrailCLI.Execute()
	if err != nil {
		logutil.FatalWithStackTrace(err)
	}
}
