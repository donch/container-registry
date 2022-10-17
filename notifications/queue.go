package notifications

import (
	"time"

	"github.com/docker/distribution"
	"github.com/docker/distribution/reference"
	"github.com/docker/distribution/uuid"
	"github.com/opencontainers/go-digest"
)

// QueueBridge is a connector in between the current request and the queue.
// It holds the context that the events need per request along with the Queue
type QueueBridge struct {
	source  SourceRecord
	actor   ActorRecord
	request RequestRecord
	// TODO: replace sink with a Queue when we implement
	// https://gitlab.com/gitlab-org/container-registry/-/issues/335
	sink Sink
}

// NewQueueBridge creates an instance of the QueueBridge with the source, actor, request and the queue
func NewQueueBridge(source SourceRecord, actor ActorRecord, request RequestRecord, sink Sink) *QueueBridge {
	return &QueueBridge{
		source:  source,
		actor:   actor,
		request: request,
		// TODO: use a queue instead of using a sink explicitly https://gitlab.com/gitlab-org/container-registry/-/issues/764
		sink: sink,
	}
}

// ManifestDeleted creates and queues an event with the deleted manifest repository and digest.
func (qb *QueueBridge) ManifestDeleted(repo reference.Named, dgst digest.Digest) error {
	event := Event{
		ID:        uuid.Generate().String(),
		Timestamp: time.Now(),
		Action:    EventActionDelete,
		Source:    qb.source,
		Actor:     qb.actor,
		Request:   qb.request,
		Target: Target{
			Descriptor: distribution.Descriptor{
				Digest: dgst,
			},
			Repository: repo.Name(),
		},
	}

	return qb.sink.Write(event)
}
