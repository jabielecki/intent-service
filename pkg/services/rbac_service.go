package services

import (
	"context"

	"github.com/tungstenfabric-preview/intent-service/pkg/auth"
	"github.com/tungstenfabric-preview/intent-service/pkg/models"
	"github.com/tungstenfabric-preview/intent-service/pkg/services/baseservices"
)

func (service *RBACService) getAllAPIAccessLists(ctx context.Context) []*models.APIAccessList {
	noAuthCtx := auth.NoAuth(ctx)
	listRequest := &ListAPIAccessListRequest{
		Spec: &baseservices.ListSpec{},
	}
	// Use a context with No auth for internal calls
	result, err := service.ReadService.ListAPIAccessList(noAuthCtx, listRequest)
	if err != nil {
		return nil
	}
	return result.APIAccessLists
}
