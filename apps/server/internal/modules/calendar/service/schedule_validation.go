package calendar

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

func validateScheduleRange(startAt, endAt time.Time) error {
	if startAt.IsZero() || endAt.IsZero() || !endAt.After(startAt) {
		return ErrInvalidScheduleRange
	}
	if endAt.Sub(startAt) > 93*24*time.Hour {
		return ErrInvalidScheduleRange
	}
	return nil
}

func normalizeScheduleBlockInput(input CoreScheduleBlockInput, now time.Time) (CoreScheduleBlockInput, error) {
	if input.WorkspaceID == uuid.Nil || input.UserID == uuid.Nil {
		return CoreScheduleBlockInput{}, ErrInvalidScheduleBlock
	}
	if err := validateScheduleRange(input.StartAt, input.EndAt); err != nil {
		return CoreScheduleBlockInput{}, err
	}
	if input.StartAt.Before(now.Add(defaultSyncLookback)) || input.EndAt.After(now.Add(defaultSyncLookahead)) {
		return CoreScheduleBlockInput{}, ErrInvalidScheduleRange
	}
	input.Title = strings.TrimSpace(input.Title)
	if input.Title == "" {
		return CoreScheduleBlockInput{}, ErrInvalidScheduleBlock
	}
	switch input.BlockType {
	case ScheduleBlockTypeWork:
		if input.StoryID == nil || *input.StoryID == uuid.Nil {
			return CoreScheduleBlockInput{}, ErrInvalidScheduleBlock
		}
	case ScheduleBlockTypeFocus:
		input.StoryID = nil
	default:
		return CoreScheduleBlockInput{}, ErrInvalidScheduleBlock
	}
	if input.Source == "" {
		input.Source = ScheduleBlockSourceUser
	}
	switch input.Source {
	case ScheduleBlockSourceUser, ScheduleBlockSourceMaya:
	default:
		return CoreScheduleBlockInput{}, ErrInvalidScheduleBlock
	}
	input.StartAt = input.StartAt.UTC()
	input.EndAt = input.EndAt.UTC()
	return input, nil
}

type randReader struct {
	read func([]byte) (int, error)
}

func (r randReader) Read(p []byte) (int, error) {
	return r.read(p)
}
