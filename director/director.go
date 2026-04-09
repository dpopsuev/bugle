package director

import (
	"context"

	troupe "github.com/dpopsuev/troupe"
)

// Director is the consumer contract for orchestration strategies.
// Origami implements CircuitDirector. Djinn implements LocalDirector.
// Directors compose: an outer Director can wrap an inner Director.
type Director interface {
	// Direct executes the orchestration plan using actors from the Broker.
	// Returns a channel of Events that streams progress until completion.
	// The channel is closed when the Director is done.
	Direct(ctx context.Context, broker troupe.Broker) (<-chan troupe.Event, error)
}
