package storieshttp

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	stories "github.com/complexus-tech/projects-api/internal/modules/stories/service"
)

func TestStoryReadStatus(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{
			name: "not found",
			err:  fmt.Errorf("load story: %w", stories.ErrNotFound),
			want: http.StatusNotFound,
		},
		{
			name: "invalid reference",
			err:  fmt.Errorf("parse story: %w", stories.ErrInvalidStoryReference),
			want: http.StatusBadRequest,
		},
		{
			name: "unexpected failure",
			err:  errors.New("database unavailable"),
			want: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := storyReadStatus(tt.err); got != tt.want {
				t.Fatalf("storyReadStatus() = %d, want %d", got, tt.want)
			}
		})
	}
}
