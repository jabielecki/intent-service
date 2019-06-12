package compilation

import (
	"github.com/tungstenfabric-preview/intent-service/pkg/apisrv/client"
	"github.com/tungstenfabric-preview/intent-service/pkg/compilation/config"
	"github.com/tungstenfabric-preview/intent-service/pkg/keystone"
)

func newAPIClient(config config.Config) *client.HTTP {
	c := config.APIClientConfig
	restClient := client.NewHTTP(
		c.URL,
		c.AuthURL,
		c.ID,
		c.Password,
		c.Insecure,
		keystone.NewScope(
			c.DomainID, c.DomainName, c.ProjectID, c.ProjectName,
		),
	)
	restClient.Init()

	return restClient
}
