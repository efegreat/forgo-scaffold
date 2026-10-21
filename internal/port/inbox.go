package port

import "context"

// FailedMessages are the messages that a subscriber failed, which the consumer's inbox keeps as failed, and
// rejects again on every redelivery.
type FailedMessages interface {
	// Retry sets the failed messages of subscriber back to waiting, with no attempts, so that the consumer
	// handles them when they are delivered again. With a messageID, it sets back that message only. It
	// returns how many it set back.
	Retry(ctx context.Context, subscriber, messageID string) (int64, error)
}
