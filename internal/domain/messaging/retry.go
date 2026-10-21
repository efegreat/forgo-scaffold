// Package messaging operates on the messages the consumers handle. Its retry-failed command retries the
// messages that a subscriber failed, after the cause of the failure was fixed.
package messaging

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/efegreat/forgo/cli"

	"example/internal/port"
)

var (
	errNoSubscriber = errors.New("retry-failed: -subscriber is required")
	errArgs         = errors.New("retry-failed: takes flags only")
)

// RetryCommand sets the failed messages of a subscriber back to waiting, the first step of retrying them.
// The consumer's inbox rejects a failed message on every redelivery, so a message replayed first would be
// rejected again, straight back to the dead-letter queue. The second step, replaying the messages from the
// dead-letter queue, is the broker's, such as RabbitMQ's management UI or a shovel.
type RetryCommand struct {
	failed     port.FailedMessages
	out        io.Writer
	flags      *flag.FlagSet
	subscriber string
	messageID  string
}

// NewRetryCommand returns the command's factory, for cli.As.
func NewRetryCommand(failed port.FailedMessages) cli.CommandFactory {
	return func() cli.Command { return &RetryCommand{failed: failed, out: os.Stdout} }
}

func (c *RetryCommand) Init(fs *flag.FlagSet) {
	c.flags = fs
	fs.StringVar(&c.subscriber, "subscriber", "", "the subscriber whose failed messages to retry, such as shipping (required)")
	fs.StringVar(&c.messageID, "message", "", "the ID of the one message to retry; every failed message of the subscriber when empty")
}

func (c *RetryCommand) Validate() error {
	if c.flags.NArg() > 0 {
		return errArgs
	}
	if c.subscriber == "" {
		return errNoSubscriber
	}
	return nil
}

func (c *RetryCommand) Run(ctx context.Context) error {
	n, err := c.failed.Retry(ctx, c.subscriber, c.messageID)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(c.out, "%d failed messages of %s set back to waiting: replay them from its dead-letter queue\n", n, c.subscriber)
	return err
}
