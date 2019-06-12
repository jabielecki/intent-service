package logic

import (
	"context"

	"github.com/tungstenfabric-preview/intent-service/pkg/compilation/intent"
	"github.com/tungstenfabric-preview/intent-service/pkg/db"
	"github.com/tungstenfabric-preview/intent-service/pkg/models"
	"github.com/tungstenfabric-preview/intent-service/pkg/services"
)

// TODO: get_autonomous_system method
const (
	defaultAutonomousSystem = 64512
	routeTargetIntPoolID    = "route_target_number"
)

func createDefaultRouteTarget(
	ctx context.Context,
	evaluateContext *intent.EvaluateContext,
) (*models.RouteTarget, error) {
	target, err := evaluateContext.IntPoolAllocator.AllocateInt(ctx, routeTargetIntPoolID, db.EmptyIntOwner)
	if err != nil {
		return nil, err
	}

	rtKey := models.RouteTargetString(defaultAutonomousSystem, target)

	rtResponse, err := evaluateContext.WriteService.CreateRouteTarget(
		ctx,
		&services.CreateRouteTargetRequest{
			RouteTarget: &models.RouteTarget{
				FQName:      []string{rtKey},
				DisplayName: rtKey,
			},
		},
	)

	return rtResponse.GetRouteTarget(), err
}
