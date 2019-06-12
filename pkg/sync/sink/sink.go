package sink

import (
	"context"

	"github.com/tungstenfabric-preview/intent-service/pkg/db"
)

// Sink represents service that handler transfers data to.
type Sink interface {
	Create(ctx context.Context, resourceName string, pk string, obj db.Object) error
	Update(ctx context.Context, resourceName string, pk string, obj db.Object) error
	Delete(ctx context.Context, resourceName string, pk string) error
}
