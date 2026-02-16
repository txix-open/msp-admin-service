package session_worker

import (
	"context"

	"github.com/pkg/errors"
	"github.com/txix-open/bgjob"
	"github.com/txix-open/isp-kit/bgjobx"
)

const (
	QueueName = "set_session_expired"
)

func EnqueueSeedJob(ctx context.Context, client *bgjobx.Client) error {
	err := client.Enqueue(ctx, bgjob.EnqueueRequest{
		Id:    "set_session_expired",
		Queue: QueueName,
		Type:  "set_session_expired",
	})
	if err != nil && !errors.Is(err, bgjob.ErrJobAlreadyExist) {
		return errors.WithMessage(err, "enqueue job")
	}

	return nil
}
