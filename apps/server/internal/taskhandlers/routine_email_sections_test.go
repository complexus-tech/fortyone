package taskhandlers

import (
	"fmt"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/stretchr/testify/require"
)

func TestRoutineSummaryUsesLocalWeekAndMorningDelivery(t *testing.T) {
	for _, test := range []struct {
		now, zone, date, week string
		ready                 bool
	}{
		{"2026-10-04T23:30:00Z", "Africa/Harare", "2026-10-05", "2026-10-05", false},
		{"2026-10-05T07:00:00Z", "Africa/Harare", "2026-10-05", "2026-10-05", true},
		{"2026-10-06T02:30:00Z", "Asia/Kolkata", "2026-10-06", "2026-10-05", false},
		{"2026-10-07T14:00:00Z", "America/New_York", "2026-10-07", "2026-10-05", true},
		{"2026-11-01T14:00:00Z", "America/New_York", "2026-11-01", "2026-10-26", true},
		{"2026-10-05T08:59:00Z", "invalid", "2026-10-05", "2026-10-05", false},
		{"2027-01-01T09:00:00Z", "UTC", "2027-01-01", "2026-12-28", true},
		{"2026-10-04T23:30:00Z", "Asia/Tokyo", "2026-10-05", "2026-10-05", false},
		{"2026-10-04T23:30:00Z", "America/New_York", "2026-10-04", "2026-09-28", true},
	} {
		t.Run(test.zone+"/"+test.now, func(t *testing.T) {
			now, err := time.Parse(time.RFC3339, test.now)
			require.NoError(t, err)
			date, ready := routineDeliveryDate(now, test.zone)
			require.Equal(t, test.date, date.Format(time.DateOnly))
			require.Equal(t, test.ready, ready)
			require.Equal(t, test.week, routineWeekStart(date).Format(time.DateOnly))
			require.Equal(t, "routine:weekly:"+test.week, routineDeliveryKey(date))
		})
	}
}

func TestRoutineSummaryCapsDetailsAcrossActivityGuidanceAndFeedback(t *testing.T) {
	sections := make([]mailer.Digest, 3)
	for index := range sections {
		sections[index].Intro = fmt.Sprintf("Section %d", index)
		for detail := range 5 {
			sections[index].Rows = append(sections[index].Rows, mailer.DigestRow{Text: fmt.Sprintf("Detail %d", detail)})
		}
		sections[index].Rows = append(sections[index].Rows, mailer.DigestRow{More: true, Text: "3 more", URL: "https://product.fortyone.app/notifications"})
	}
	bounded := limitRoutineEmailSections(sections, "https://product.fortyone.app")
	details := 0
	for _, section := range bounded {
		for _, row := range section.Rows {
			if !row.More {
				details++
			}
		}
		require.True(t, section.Rows[len(section.Rows)-1].More)
	}
	require.Equal(t, 10, details)
	require.Len(t, bounded[2].Rows, 1)
	require.Len(t, sections[2].Rows, 6)
}
