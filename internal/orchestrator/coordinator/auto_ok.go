package coordinator

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/version-1/tasq/internal/issue/domain/entity"
)

type autoOKTracker interface {
	CompleteAutoOK(context.Context, int64, entity.AutoOKResultInput) (entity.Issue, error)
}

func completeAutoOK(ctx context.Context, tracker any, issue entity.Issue, deliveryErr error) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	client, ok := tracker.(autoOKTracker)
	if !ok {
		return errors.New("tracker does not support Auto OK results")
	}
	input := entity.AutoOKResultInput{Generation: issue.AutoOKGeneration}
	if deliveryErr != nil {
		input.Error = deliveryErr.Error()
	}
	_, err := client.CompleteAutoOK(ctx, issue.ID, input)
	if err != nil {
		return fmt.Errorf("record Auto OK result for issue %d: %w", issue.ID, err)
	}
	return nil
}
