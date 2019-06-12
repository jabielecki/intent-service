package integration

import (
	"context"

	"github.com/tungstenfabric-preview/intent-service/pkg/constants"
	"github.com/tungstenfabric-preview/intent-service/pkg/db/cache"
	"github.com/tungstenfabric-preview/intent-service/pkg/db/etcd"
	integrationetcd "github.com/tungstenfabric-preview/intent-service/pkg/testutil/integration/etcd"
)

const (
	maxHistory = 100000
)

// RunCacheDB runs DB Cache with etcd event producer.
func RunCacheDB() (*cache.DB, func() error, error) {
	setViperConfig(map[string]interface{}{
		"cache.timeout":           "10s",
		constants.ETCDEndpointsVK: []string{integrationetcd.Endpoint},
	})

	cacheDB := cache.NewDB(maxHistory)

	processor, err := etcd.NewEventProducer(cacheDB, "integration-cache-db")
	if err != nil {
		return nil, func() error { return nil }, err
	}

	ctx, cancelEtcdEventProducer := context.WithCancel(context.Background())
	errChan := make(chan error)
	go func() {
		errChan <- processor.Start(ctx)
	}()

	return cacheDB, func() error {
		cancelEtcdEventProducer()
		return <-errChan
	}, nil
}
