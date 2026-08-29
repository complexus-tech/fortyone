package objectiveshttp

import (
	"encoding/json"
	"strings"
	"testing"

	objectives "github.com/complexus-tech/projects-api/internal/modules/objectives/service"
	"github.com/google/uuid"
)

func TestToAppStrategyMapSerializesEmptyObjectiveIDsAsArray(t *testing.T) {
	strategy := toAppStrategyMap(objectives.CoreStrategyMap{
		Pillars: []objectives.CoreStrategicPillar{
			{
				ID:           uuid.New(),
				Name:         "Customer trust",
				ObjectiveIDs: nil,
			},
		},
	})

	payload, err := json.Marshal(strategy)
	if err != nil {
		t.Fatalf("marshal strategy map: %v", err)
	}
	if !strings.Contains(string(payload), `"objectiveIds":[]`) {
		t.Fatalf("expected objectiveIds to be an array, got %s", payload)
	}
}
