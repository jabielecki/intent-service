package types

import (
	"context"

	"github.com/tungstenfabric-preview/intent-service/pkg/services"
	"github.com/tungstenfabric-preview/intent-service/pkg/types/ipam"
)

const (
	// VirtualNetworkIDPoolKey is a key for id pool for virtual network id.
	VirtualNetworkIDPoolKey = "virtual_network_id"
)

// InTransactionDoer makes transaction mocking possible in type logic tests
type InTransactionDoer interface {
	DoInTransaction(ctx context.Context, do func(context.Context) error) error
}

// ContrailTypeLogicService is a service for implementing type specific logic
type ContrailTypeLogicService struct {
	services.BaseService
	ReadService       services.ReadService
	InTransactionDoer InTransactionDoer
	AddressManager    ipam.AddressManager
	IntPoolAllocator  ipam.IntPoolAllocator
	WriteService      services.WriteService
}
