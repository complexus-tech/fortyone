package notifications

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestNotificationMessageRemainsBackwardCompatibleWithoutStrategySnapshot(t *testing.T) {
	legacyJSON := []byte(`{
		"template":"{actor} updated the objective",
		"variables":{"actor":{"value":"Joseph","type":"actor"}}
	}`)

	var message NotificationMessage
	require.NoError(t, json.Unmarshal(legacyJSON, &message))
	require.Nil(t, message.Strategy)

	encoded, err := json.Marshal(message)
	require.NoError(t, err)
	require.NotContains(t, string(encoded), `"strategy"`)
}

func TestNotificationMessageRoundTripsStructuredStrategySnapshot(t *testing.T) {
	generatedAt := time.Date(2026, time.August, 12, 7, 0, 0, 0, time.UTC)
	updatedAt := generatedAt.Add(-8 * 24 * time.Hour)
	health := "At Risk"
	message := NotificationMessage{
		Template:  "Weekly strategy check-in",
		Variables: map[string]Variable{},
		Strategy: &StrategyNotificationSnapshot{
			Version:     1,
			Kind:        StrategyNotificationKindWeeklyCheckIn,
			GeneratedAt: generatedAt,
			WeeklyCheckIn: &StrategyWeeklyCheckInSnapshot{
				StaleAfterDays: 7,
				Counts: StrategyWeeklyCheckInCounts{
					AtRiskObjectives: 1,
					StaleObjectives:  1,
					UniqueObjectives: 1,
				},
				Objectives: []StrategyObjectiveSnapshot{
					{
						ID:        uuid.MustParse("10000000-0000-0000-0000-000000000001"),
						TeamID:    uuid.MustParse("20000000-0000-0000-0000-000000000001"),
						Name:      "Grow enterprise revenue",
						Health:    &health,
						UpdatedAt: updatedAt,
						Reasons: []string{
							StrategySignalReasonAtRisk,
							StrategySignalReasonStale,
						},
					},
				},
				KeyResults: []StrategyKeyResultSnapshot{},
				OmittedDetails: &StrategyWeeklyCheckInOmittedDetailsSnapshot{
					Objectives: 2,
					KeyResults: 3,
				},
			},
		},
	}

	encoded, err := json.Marshal(message)
	require.NoError(t, err)

	var decoded NotificationMessage
	require.NoError(t, json.Unmarshal(encoded, &decoded))
	require.Equal(t, message, decoded)
}

func TestMonthlyStrategySnapshotDistinguishesNoKeyResultsAndPreservesLegacyProgress(t *testing.T) {
	message := NotificationMessage{
		Template:  "Monthly strategy summary",
		Variables: map[string]Variable{},
		Strategy: &StrategyNotificationSnapshot{
			Version: 1,
			Kind:    StrategyNotificationKindMonthlySummary,
			MonthlySummary: &StrategyMonthlySummarySnapshot{
				PeriodStart: time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC),
				PeriodEnd:   time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	encoded, err := json.Marshal(message)
	require.NoError(t, err)
	require.Contains(t, string(encoded), `"keyResultCount":0`)
	require.NotContains(t, string(encoded), `"keyResultProgress"`)

	var legacy NotificationMessage
	require.NoError(t, json.Unmarshal([]byte(`{
		"template":"Monthly strategy summary",
		"variables":{},
		"strategy":{
			"version":1,
			"kind":"monthly_summary",
			"generatedAt":"2026-08-01T09:00:00Z",
			"monthlySummary":{
				"periodStart":"2026-07-01T00:00:00Z",
				"periodEnd":"2026-08-01T00:00:00Z",
				"keyResultProgress":46
			}
		}
	}`), &legacy))
	require.NotNil(t, legacy.Strategy.MonthlySummary.KeyResultProgress)
	require.Equal(t, 46.0, *legacy.Strategy.MonthlySummary.KeyResultProgress)
}

func TestCoreNotificationPublicRedactsStrategyWithoutMutatingSource(t *testing.T) {
	source := CoreNotification{
		EntityType: "strategy",
		Message: NotificationMessage{
			Template: "Your weekly strategy check-in",
			Variables: map[string]Variable{
				"actor": {Value: "Maya", Type: "actor"},
			},
			Strategy: &StrategyNotificationSnapshot{
				Version: 1,
				Kind:    StrategyNotificationKindWeeklyCheckIn,
			},
		},
	}

	public := source.Public()

	require.Nil(t, public.Message.Strategy)
	require.Equal(t, "Strategy guidance is ready to review.", public.Message.Template)
	require.Empty(t, public.Message.Variables)
	require.Equal(t, "Your weekly strategy check-in", source.Message.Template)
	require.Equal(t, "Maya", source.Message.Variables["actor"].Value)
	require.NotNil(t, source.Message.Strategy)
}
