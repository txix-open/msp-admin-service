package delete_old_audit_worker

import (
	"context"

	"github.com/pkg/errors"
	"github.com/txix-open/bgjob"
	"github.com/txix-open/isp-kit/bgjobx"
)

const (
	QueueName = "delete_old_audit"
)

func EnqueueSeedJob(ctx context.Context, client *bgjobx.Client) error {
	err := client.Enqueue(ctx, bgjob.EnqueueRequest{
		Id:    "delete_old_audit",
		Queue: QueueName,
		Type:  "delete_old_audit",
	})
	if err != nil && !errors.Is(err, bgjob.ErrJobAlreadyExist) {
		return errors.WithMessage(err, "enqueue job")
	}

	return nil
}
