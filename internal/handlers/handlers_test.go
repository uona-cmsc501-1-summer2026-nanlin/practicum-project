package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gobeetle/reply"
	"github.com/gofiber/fiber/v2"
	"github.com/uona-cmsc501-1-summer2026-nanlin/practicum-project/internal/database"
)

func testApp(t *testing.T) *fiber.App {
	t.Helper()
	db, err := database.Connect(":memory:")
	if err != nil {
		t.Fatalf("database: %v", err)
	}
	app := fiber.New(fiber.Config{
		ErrorHandler: reply.FiberErrorHandler(),
	})
	api := &API{DB: db}
	api.Register(app)
	return app
}

func doJSON(t *testing.T, app *fiber.App, method, path string, body any) (int, map[string]any) {
	t.Helper()
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		reader = bytes.NewReader(raw)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	out := map[string]any{}
	if len(data) > 0 {
		_ = json.Unmarshal(data, &out)
	}
	return resp.StatusCode, out
}

func responseData(body map[string]any) map[string]any {
	d, _ := body["data"].(map[string]any)
	return d
}

func TestCreateGroup_validation(t *testing.T) {
	app := testApp(t)

	code, body := doJSON(t, app, "POST", "/api/v1/groups", map[string]string{"name": ""})
	if code != 400 {
		t.Fatalf("empty name: status=%d body=%v", code, body)
	}
	if body["code"].(float64) != 400 {
		t.Fatalf("expected reply code 400, body=%v", body)
	}

	code, body = doJSON(t, app, "POST", "/api/v1/groups", map[string]string{"name": "Trip"})
	if code != 201 {
		t.Fatalf("valid group: status=%d body=%v", code, body)
	}
	if responseData(body)["name"] != "Trip" {
		t.Fatalf("expected group in data envelope, body=%v", body)
	}
}

func TestCreateCharge_validation(t *testing.T) {
	app := testApp(t)
	_, g := doJSON(t, app, "POST", "/api/v1/groups", map[string]string{"name": "Trip"})
	groupID := uint(responseData(g)["id"].(float64))

	_, p := doJSON(t, app, "POST", "/api/v1/groups/"+gid(groupID)+"/people", map[string]string{"name": "Alex"})
	personID := uint(responseData(p)["id"].(float64))

	code, body := doJSON(t, app, "POST", "/api/v1/groups/"+gid(groupID)+"/charges", map[string]any{
		"description":    "Dinner",
		"amount":         0,
		"paidByPersonId": personID,
		"participantIds": []uint{personID},
	})
	if code != 400 || body["code"].(float64) != 400 {
		t.Fatalf("zero amount: status=%d body=%v", code, body)
	}

	code, body = doJSON(t, app, "POST", "/api/v1/groups/"+gid(groupID)+"/charges", map[string]any{
		"description":    "Dinner",
		"amount":         90,
		"paidByPersonId": 999,
		"participantIds": []uint{personID},
	})
	if code != 404 || body["code"].(float64) != 404 {
		t.Fatalf("bad payer: status=%d body=%v", code, body)
	}
}

func TestSettle_happyPath(t *testing.T) {
	app := testApp(t)
	_, g := doJSON(t, app, "POST", "/api/v1/groups", map[string]string{"name": "July trip"})
	groupID := uint(responseData(g)["id"].(float64))
	path := gid(groupID)

	_, p1 := doJSON(t, app, "POST", "/api/v1/groups/"+path+"/people", map[string]string{"name": "Alex"})
	_, p2 := doJSON(t, app, "POST", "/api/v1/groups/"+path+"/people", map[string]string{"name": "Sam"})
	_, p3 := doJSON(t, app, "POST", "/api/v1/groups/"+path+"/people", map[string]string{"name": "Jordan"})
	id1 := uint(responseData(p1)["id"].(float64))
	id2 := uint(responseData(p2)["id"].(float64))
	id3 := uint(responseData(p3)["id"].(float64))

	doJSON(t, app, "POST", "/api/v1/groups/"+path+"/charges", map[string]any{
		"description":    "Dinner",
		"amount":         90,
		"paidByPersonId": id1,
		"participantIds": []uint{id1, id2, id3},
	})

	code, body := doJSON(t, app, "POST", "/api/v1/groups/"+path+"/settle", map[string]any{})
	if code != 200 {
		t.Fatalf("settle: status=%d body=%v", code, body)
	}
	settle := responseData(body)
	balances := settle["balances"].([]any)
	if len(balances) != 3 {
		t.Fatalf("expected 3 balances, got %v", balances)
	}
}

func TestSettle_missingGroup(t *testing.T) {
	app := testApp(t)
	code, body := doJSON(t, app, "POST", "/api/v1/groups/999/settle", map[string]any{})
	if code != 404 || body["code"].(float64) != 404 {
		t.Fatalf("missing group: status=%d body=%v", code, body)
	}
}

func gid(groupID uint) string {
	return fmt.Sprintf("%d", groupID)
}
