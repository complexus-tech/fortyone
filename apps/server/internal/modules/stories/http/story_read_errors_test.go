package storieshttp

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
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
			name: "invalid typed query",
			err:  fmt.Errorf("filter story: %w", stories.ErrInvalidStoryReadQuery),
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

func TestScopedMutationStatuses(t *testing.T) {
	tests := []struct {
		name string
		got  int
		want int
	}{
		{name: "bulk delete forbidden", got: bulkDeleteStatus(stories.ErrDeleteForbidden), want: http.StatusForbidden},
		{name: "bulk delete hidden target", got: bulkDeleteStatus(stories.ErrNotFound), want: http.StatusNotFound},
		{name: "attachment forbidden", got: storyAttachmentStatus(attachments.ErrUnauthorized), want: http.StatusForbidden},
		{name: "attachment hidden target", got: storyAttachmentStatus(attachments.ErrNotFound), want: http.StatusNotFound},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.got != test.want {
				t.Fatalf("status = %d, want %d", test.got, test.want)
			}
		})
	}
}

func TestStoryMutationStatus(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{
			name: "workspace not entitled",
			err:  fmt.Errorf("enable scheduling: %w", stories.ErrAutoSchedulingUnavailable),
			want: http.StatusPaymentRequired,
		},
		{
			name: "scheduling entitlement check unavailable",
			err:  fmt.Errorf("enable scheduling: %w", stories.ErrAutoSchedulingAccessCheckFailed),
			want: http.StatusServiceUnavailable,
		},
		{
			name: "Maya scheduling contract",
			err:  fmt.Errorf("validate assignment: %w", stories.ErrMayaAssignmentRequiresDuration),
			want: http.StatusBadRequest,
		},
		{
			name: "missing story",
			err:  fmt.Errorf("delete story: %w", stories.ErrNotFound),
			want: http.StatusNotFound,
		},
		{
			name: "delete forbidden",
			err:  fmt.Errorf("delete story: %w", stories.ErrDeleteForbidden),
			want: http.StatusForbidden,
		},
		{
			name: "unknown repository failure",
			err:  errors.New("database unavailable"),
			want: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := storyMutationStatus(test.err); got != test.want {
				t.Fatalf("storyMutationStatus() = %d, want %d", got, test.want)
			}
		})
	}
}
