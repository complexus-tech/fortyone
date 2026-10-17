package nativepreviewhttp

import (
	"github.com/google/uuid"
	"testing"
)

func TestParseTarget(t *testing.T) {
	id := uuid.NewString()
	team := uuid.NewString()
	for _, test := range []struct {
		kind, id, team string
		valid          bool
	}{{"story", id, "", true}, {"story", "eng-42", "", true}, {"objective", id, team, true}, {"document", id, "", true}, {"story", "ENG-0", "", false}, {"story", "ENG-9999999999999999999", "", false}, {"objective", id, "", false}, {"document", "ENG-42", "", false}, {"unknown", id, "", false}, {"story", uuid.Nil.String(), "", false}} {
		t.Run(test.kind+test.id+test.team, func(t *testing.T) {
			_, err := parseTarget(test.kind, test.id, test.team)
			if (err == nil) != test.valid {
				t.Fatalf("err=%v valid=%v", err, test.valid)
			}
		})
	}
}
