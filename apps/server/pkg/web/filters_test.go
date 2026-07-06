package web

import (
	"net/url"
	"testing"
)

func TestGetFiltersUsesJSONTagWhenDBTagIsMissing(t *testing.T) {
	type filters struct {
		Page     int `json:"page"`
		PageSize int `json:"pageSize"`
	}

	result, err := GetFilters(url.Values{
		"page":     []string{"2"},
		"pageSize": []string{"15"},
	}, &filters{})
	if err != nil {
		t.Fatalf("GetFilters returned error: %v", err)
	}

	if _, ok := result[""]; ok {
		t.Fatal("expected filters not to contain an empty key")
	}
	if result["page"] != 2 {
		t.Fatalf("expected page filter 2, got %#v", result["page"])
	}
	if result["pageSize"] != 15 {
		t.Fatalf("expected pageSize filter 15, got %#v", result["pageSize"])
	}
}
