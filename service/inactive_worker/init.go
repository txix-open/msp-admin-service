package inactive_worker

import (
	"context"

	"github.com/integration-system/bgjob"
	"github.com/integration-system/isp-kit/bgjobx"
	"github.com/pkg/errors"
)

const (
	QueueName = "block_inactive_user"
)

func EnqueueSeedJob(ctx context.Context, client *bgjobx.Client) error {
	err := client.Enqueue(ctx, bgjob.EnqueueRequest{
		Id:    "block_inactive_user",
		Queue: QueueName,
		Type:  "block_inactive_user",
	})
	if err != nil && !errors.Is(err, bgjob.ErrJobAlreadyExist) {
		return errors.WithMessage(err, "enqueue job")
	}

	return nil
}
