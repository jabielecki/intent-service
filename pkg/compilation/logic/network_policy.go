package logic

import (
	"github.com/tungstenfabric-preview/intent-service/pkg/compilation/intent"
	"github.com/tungstenfabric-preview/intent-service/pkg/models"
	"github.com/tungstenfabric-preview/intent-service/pkg/models/basemodels"
)

// NetworkPolicyIntent intent
type NetworkPolicyIntent struct {
	intent.BaseIntent
	*models.NetworkPolicy
}

// GetObject returns embedded resource object
func (i *NetworkPolicyIntent) GetObject() basemodels.Object {
	return i.NetworkPolicy
}
