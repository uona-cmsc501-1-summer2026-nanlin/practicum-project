package handlers

import (
	"testing"
)

func TestTruncateAll_clearsData(t *testing.T) {
	app := testApp(t)

	doJSON(t, app, "POST", "/api/v1/groups", map[string]string{"name": "Trip"})
	doJSON(t, app, "POST", "/api/v1/groups/1/people", map[string]string{"name": "Alex"})
	doJSON(t, app, "POST", "/api/v1/groups/1/charges", map[string]any{
		"description":    "Dinner",
		"amount":         30.0,
		"paidByPersonId": 1,
		"participantIds": []int{1},
	})

	code, body := doJSON(t, app, "DELETE", "/api/v1/internal/testing/truncate/all", nil)
	if code != 200 {
		t.Fatalf("truncate all: status=%d body=%v", code, body)
	}
	data := responseData(body)
	if data["groups"].(float64) != 1 {
		t.Fatalf("expected 1 group deleted, got %v", data)
	}
	if data["people"].(float64) != 1 {
		t.Fatalf("expected 1 person deleted, got %v", data)
	}
	if data["charges"].(float64) != 1 {
		t.Fatalf("expected 1 charge deleted, got %v", data)
	}

	code, list := doJSON(t, app, "GET", "/api/v1/groups", nil)
	if code != 200 {
		t.Fatalf("list groups: status=%d", code)
	}
	groups, _ := list["data"].([]any)
	if len(groups) != 0 {
		t.Fatalf("expected empty groups after truncate, got %v", groups)
	}
}

func TestTruncateCharges_onlyCharges(t *testing.T) {
	app := testApp(t)

	doJSON(t, app, "POST", "/api/v1/groups", map[string]string{"name": "Trip"})
	doJSON(t, app, "POST", "/api/v1/groups/1/people", map[string]string{"name": "Alex"})
	doJSON(t, app, "POST", "/api/v1/groups/1/charges", map[string]any{
		"description":    "Dinner",
		"amount":         30.0,
		"paidByPersonId": 1,
		"participantIds": []int{1},
	})

	code, body := doJSON(t, app, "DELETE", "/api/v1/internal/testing/truncate/charges", nil)
	if code != 200 {
		t.Fatalf("truncate charges: status=%d body=%v", code, body)
	}
	data := responseData(body)
	if data["charges"].(float64) != 1 {
		t.Fatalf("expected 1 charge deleted, got %v", data)
	}
	if data["people"].(float64) != 0 || data["groups"].(float64) != 0 {
		t.Fatalf("expected only charges deleted, got %v", data)
	}

	code, list := doJSON(t, app, "GET", "/api/v1/groups/1/people", nil)
	if code != 200 {
		t.Fatalf("list people: status=%d", code)
	}
	people, _ := list["data"].([]any)
	if len(people) != 1 {
		t.Fatalf("expected people to remain, got %v", people)
	}
}
