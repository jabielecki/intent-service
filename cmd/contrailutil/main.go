package main

import (
	"github.com/tungstenfabric-preview/intent-service/pkg/cmd/contrailutil"
	"github.com/tungstenfabric-preview/intent-service/pkg/logutil"
)

func main() {
	err := contrailutil.ContrailUtil.Execute()
	if err != nil {
		logutil.FatalWithStackTrace(err)
	}
}
