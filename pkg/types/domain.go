package types

import (
	"context"

	"github.com/tungstenfabric-preview/intent-service/pkg/models"
	"github.com/tungstenfabric-preview/intent-service/pkg/models/basemodels"
	"github.com/tungstenfabric-preview/intent-service/pkg/services"
)

//CreateDomain does pre-check for create domain
func (sv *ContrailTypeLogicService) CreateDomain(
	ctx context.Context,
	request *services.CreateDomainRequest,
) (response *services.CreateDomainResponse, err error) {
	domain := request.GetDomain()
	err = sv.InTransactionDoer.DoInTransaction(
		ctx,
		func(ctx context.Context) error {
			domain.Perms2.Share = append(domain.Perms2.Share, &models.ShareType{
				TenantAccess: basemodels.PermsRW,
				Tenant:       "domain:" + domain.UUID,
			})
			response, err = sv.BaseService.CreateDomain(ctx, request)
			return err
		})
	return response, err
}
