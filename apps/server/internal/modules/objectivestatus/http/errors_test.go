package objectivestatushttp

import (
	"errors"
	"net/http"
	"testing"

	objectivestatus "github.com/complexus-tech/projects-api/internal/modules/objectivestatus/service"
)

func TestObjectiveStatusErrorStatus(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		err  error
		want int
	}{
		{err: objectivestatus.ErrNotFound, want: http.StatusNotFound},
		{err: objectivestatus.ErrNoFields, want: http.StatusBadRequest},
		{err: objectivestatus.ErrInvalidOrder, want: http.StatusBadRequest},
		{err: objectivestatus.ErrStatusHasObjectives, want: http.StatusConflict},
		{err: objectivestatus.ErrLastInCategory, want: http.StatusConflict},
		{err: errors.New("database unavailable"), want: http.StatusInternalServerError},
	} {
		if got := objectiveStatusErrorStatus(test.err); got != test.want {
			t.Fatalf("objectiveStatusErrorStatus(%v) = %d, want %d", test.err, got, test.want)
		}
	}
}
