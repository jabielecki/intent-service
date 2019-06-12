package replication

import (
	"github.com/tungstenfabric-preview/intent-service/pkg/services"
)

// EventDecoder is capable of decoding row data in form of map into an Event.
type EventDecoder interface {
	DecodeRowEvent(operation, resourceName string, pk []string, properties map[string]interface{}) (*services.Event, error)
}
