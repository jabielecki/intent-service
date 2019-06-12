package main

import (
	"github.com/tungstenfabric-preview/intent-service/pkg/cmd/contrail"
	"github.com/tungstenfabric-preview/intent-service/pkg/logutil"
)

func main() {
	err := contrail.Contrail.Execute()
	if err != nil {
		logutil.FatalWithStackTrace(err)
	}
}
