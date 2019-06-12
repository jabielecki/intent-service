package types

import (
	"context"

	"github.com/tungstenfabric-preview/intent-service/pkg/auth"
	"github.com/tungstenfabric-preview/intent-service/pkg/errutil"
)

type draftModeStateGetter interface {
	GetDraftModeState() string
}

// DraftModeStateChecker checks if request contains draftModeState property
type DraftModeStateChecker interface {
	CheckDraftModeState(context.Context, draftModeStateGetter) error
}

func checkDraftModeState(ctx context.Context, dms draftModeStateGetter) error {
	if auth.IsInternalRequest(ctx) {
		return nil
	}

	if dms.GetDraftModeState() != "" {
		return errutil.ErrorBadRequest(
			"security resource property 'draft_mode_state' is only readable",
		)
	}

	return nil
}
