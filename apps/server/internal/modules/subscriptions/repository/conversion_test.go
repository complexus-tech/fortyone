package subscriptionsrepository

import (
	"errors"
	"math"
	"testing"

	subscriptionsdomain "github.com/complexus-tech/projects-api/internal/modules/subscriptions/domain"
)

func TestSubscriptionSeatCountRejectsOverflow(t *testing.T) {
	t.Parallel()

	for _, value := range []int{-1, int(int64(math.MaxInt32) + 1)} {
		_, err := subscriptionSeatCount(value)
		if !errors.Is(err, subscriptionsdomain.ErrInvalidSeatCount) {
			t.Fatalf("subscriptionSeatCount(%d) error = %v, want ErrInvalidSeatCount", value, err)
		}
	}
}
