package delete_old_audit_worker

import (
	"context"

	"github.com/integration-system/bgjob"
	"github.com/integration-system/isp-kit/bgjobx"
	"github.com/pkg/errors"
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
