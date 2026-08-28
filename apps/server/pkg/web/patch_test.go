package web_test

import (
	"encoding/json"
	"testing"

	"github.com/complexus-tech/projects-api/pkg/web"
)

func TestPatchFieldDistinguishesOmittedNullAndZero(t *testing.T) {
	t.Parallel()

	var request struct {
		Omitted web.PatchField[int] `json:"omitted"`
		Cleared web.PatchField[int] `json:"cleared"`
		Zero    web.PatchField[int] `json:"zero"`
	}
	if err := json.Unmarshal([]byte(`{"cleared":null,"zero":0}`), &request); err != nil {
		t.Fatalf("decode patch fields: %v", err)
	}

	if _, specified := request.Omitted.Value(); specified {
		t.Fatal("omitted field was marked specified")
	}
	if value, specified := request.Cleared.Value(); !specified || value != nil {
		t.Fatalf("cleared field = (%v, %t), want (nil, true)", value, specified)
	}
	if value, specified := request.Zero.Value(); !specified || value == nil || *value != 0 {
		t.Fatalf("zero field = (%v, %t), want (0, true)", value, specified)
	}
}
