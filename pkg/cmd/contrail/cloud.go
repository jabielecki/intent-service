package contrail

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/tungstenfabric-preview/intent-service/pkg/cloud"
)

func init() {
	Contrail.AddCommand(cloudCmd)
}

var cloudCmd = &cobra.Command{
	Use:   "cloud",
	Short: "sub command cloud is used to manage public cloud infra",
	Long: `Cloud is a sub command used to manage
            public cloud infra. Currently
            supported infra are Azure`,
	Run: func(cmd *cobra.Command, args []string) {
		manageCloud()
	},
}

func manageCloud() {
	manager, err := cloud.NewCloudManager(configFile)
	if err != nil {
		log.Fatal(err)
	}

	if err := manager.Manage(); err != nil {
		log.Fatal(err)
	}
}
