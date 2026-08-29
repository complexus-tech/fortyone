package stateshttp

import (
	"errors"
	"net/http"
	"testing"

	states "github.com/complexus-tech/projects-api/internal/modules/states/service"
)

func TestStateErrorStatus(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		err  error
		want int
	}{
		{err: states.ErrNotFound, want: http.StatusNotFound},
		{err: states.ErrNoFields, want: http.StatusBadRequest},
		{err: states.ErrInvalidOrder, want: http.StatusBadRequest},
		{err: states.ErrStatusHasStories, want: http.StatusConflict},
		{err: states.ErrLastInCategory, want: http.StatusConflict},
		{err: errors.New("database unavailable"), want: http.StatusInternalServerError},
	} {
		if got := stateErrorStatus(test.err); got != test.want {
			t.Fatalf("stateErrorStatus(%v) = %d, want %d", test.err, got, test.want)
		}
	}
}
