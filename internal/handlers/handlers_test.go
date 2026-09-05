package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"strings"
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

func gid(groupID uint) string {
	return fmt.Sprintf("%d", groupID)
}

func seedTripWithMembers(t *testing.T, app *fiber.App) (groupID, id1, id2, id3 uint) {
	t.Helper()
	_, g := doJSON(t, app, "POST", "/api/v1/groups", map[string]string{"name": "July trip"})
	groupID = uint(responseData(g)["id"].(float64))

	_, p1 := doJSON(t, app, "POST", "/api/v1/people", map[string]string{"name": "Alex"})
	_, p2 := doJSON(t, app, "POST", "/api/v1/people", map[string]string{"name": "Sam"})
	_, p3 := doJSON(t, app, "POST", "/api/v1/people", map[string]string{"name": "Jordan"})
	id1 = uint(responseData(p1)["id"].(float64))
	id2 = uint(responseData(p2)["id"].(float64))
	id3 = uint(responseData(p3)["id"].(float64))

	path := gid(groupID)
	doJSON(t, app, "POST", "/api/v1/groups/"+path+"/members", map[string]any{"userId": id1})
	doJSON(t, app, "POST", "/api/v1/groups/"+path+"/members", map[string]any{"userId": id2})
	doJSON(t, app, "POST", "/api/v1/groups/"+path+"/members", map[string]any{"userId": id3})
	return groupID, id1, id2, id3
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
	if responseData(body)["currency"] != "USD" {
		t.Fatalf("expected default USD, body=%v", body)
	}

	code, body = doJSON(t, app, "POST", "/api/v1/groups", map[string]string{"name": "Euro trip", "currency": "EUR"})
	if code != 201 || responseData(body)["currency"] != "EUR" {
		t.Fatalf("EUR group: status=%d body=%v", code, body)
	}

	code, body = doJSON(t, app, "POST", "/api/v1/groups", map[string]string{"name": "Bad", "currency": "XYZ"})
	if code != 400 {
		t.Fatalf("bad currency: status=%d body=%v", code, body)
	}

	code, body = doJSON(t, app, "PUT", "/api/v1/groups/1", map[string]string{"name": "July trip", "currency": "EUR"})
	if code != 200 || responseData(body)["currency"] != "EUR" || responseData(body)["name"] != "July trip" {
		t.Fatalf("update group: status=%d body=%v", code, body)
	}
}

func TestListCurrencies(t *testing.T) {
	app := testApp(t)
	code, body := doJSON(t, app, "GET", "/api/v1/currencies", nil)
	if code != 200 {
		t.Fatalf("list currencies: status=%d body=%v", code, body)
	}
	list, _ := body["data"].([]any)
	if len(list) < 1 {
		t.Fatalf("expected currency list, got %v", body)
	}
	first, _ := list[0].(map[string]any)
	if first["code"] != "USD" || first["country"] != "us" {
		t.Fatalf("unexpected first currency: %v", first)
	}
}

func TestCreatePerson_global(t *testing.T) {
	app := testApp(t)

	code, body := doJSON(t, app, "POST", "/api/v1/people", map[string]string{"name": "Alex", "email": "alex@example.com"})
	if code != 201 {
		t.Fatalf("create person: status=%d body=%v", code, body)
	}
	if responseData(body)["name"] != "Alex" {
		t.Fatalf("expected Alex, body=%v", body)
	}
	if _, ok := responseData(body)["groupId"]; ok {
		t.Fatalf("global person must not have groupId, body=%v", body)
	}

	code, list := doJSON(t, app, "GET", "/api/v1/people", nil)
	if code != 200 {
		t.Fatalf("list people: status=%d", code)
	}
	people, _ := list["data"].([]any)
	if len(people) != 1 {
		t.Fatalf("expected 1 person, got %v", people)
	}
}

func TestUpdatePerson(t *testing.T) {
	app := testApp(t)
	code, body := doJSON(t, app, "POST", "/api/v1/people", map[string]string{"name": "Alex", "email": "old@example.com"})
	if code != 201 {
		t.Fatalf("create person: status=%d body=%v", code, body)
	}
	id := uint(responseData(body)["id"].(float64))

	code, updated := doJSON(t, app, "PUT", "/api/v1/people/"+gid(id), map[string]string{
		"name":  "Alexandra",
		"email": "new@example.com",
	})
	if code != 200 {
		t.Fatalf("update person: status=%d body=%v", code, updated)
	}
	data := responseData(updated)
	if data["name"] != "Alexandra" || data["email"] != "new@example.com" {
		t.Fatalf("unexpected update result: %v", data)
	}
}

func TestAddMember_andCharge(t *testing.T) {
	app := testApp(t)
	groupID, id1, _, _ := seedTripWithMembers(t, app)
	path := gid(groupID)

	code, members := doJSON(t, app, "GET", "/api/v1/groups/"+path+"/members", nil)
	if code != 200 {
		t.Fatalf("list members: status=%d", code)
	}
	list, _ := members["data"].([]any)
	if len(list) != 3 {
		t.Fatalf("expected 3 members, got %v", list)
	}

	code, body := doJSON(t, app, "POST", "/api/v1/groups/"+path+"/charges", map[string]any{
		"description":    "Dinner",
		"amount":         0,
		"paidByUserId":   id1,
		"participantIds": []uint{id1},
	})
	if code != 400 || body["code"].(float64) != 400 {
		t.Fatalf("zero amount: status=%d body=%v", code, body)
	}

	code, body = doJSON(t, app, "POST", "/api/v1/groups/"+path+"/charges", map[string]any{
		"description":    "Dinner",
		"amount":         90,
		"paidByUserId":   999,
		"participantIds": []uint{id1},
	})
	if code != 404 || body["code"].(float64) != 404 {
		t.Fatalf("bad payer: status=%d body=%v", code, body)
	}
}

func TestSettle_happyPath(t *testing.T) {
	app := testApp(t)
	groupID, id1, id2, id3 := seedTripWithMembers(t, app)
	path := gid(groupID)

	doJSON(t, app, "POST", "/api/v1/groups/"+path+"/charges", map[string]any{
		"description":    "Dinner",
		"amount":         90,
		"paidByUserId":   id1,
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

func TestDeletePerson_blockedWhenMember(t *testing.T) {
	app := testApp(t)
	_, id1, _, _ := seedTripWithMembers(t, app)

	code, body := doJSON(t, app, "DELETE", "/api/v1/people/"+gid(id1), nil)
	if code != 400 {
		t.Fatalf("expected 400 when member, status=%d body=%v", code, body)
	}
}

func TestRemoveMember_blockedWhenOnCharge(t *testing.T) {
	app := testApp(t)
	groupID, id1, id2, id3 := seedTripWithMembers(t, app)
	path := gid(groupID)

	doJSON(t, app, "POST", "/api/v1/groups/"+path+"/charges", map[string]any{
		"description":    "Dinner",
		"amount":         90,
		"paidByUserId":   id1,
		"participantIds": []uint{id1, id2, id3},
	})

	code, body := doJSON(t, app, "DELETE", "/api/v1/groups/"+path+"/members/"+gid(id2), nil)
	if code != 400 {
		t.Fatalf("expected 400 when participant, status=%d body=%v", code, body)
	}
	if !strings.Contains(fmt.Sprint(body["error"]), "Dinner") {
		t.Fatalf("expected charge name in error, body=%v", body)
	}

	code, body = doJSON(t, app, "DELETE", "/api/v1/groups/"+path+"/members/"+gid(id1), nil)
	if code != 400 {
		t.Fatalf("expected 400 when payer, status=%d body=%v", code, body)
	}
}

func TestRemoveMember_allowedWhenNotOnCharges(t *testing.T) {
	app := testApp(t)
	groupID, id1, id2, id3 := seedTripWithMembers(t, app)
	path := gid(groupID)

	doJSON(t, app, "POST", "/api/v1/groups/"+path+"/charges", map[string]any{
		"description":    "Taxi",
		"amount":         30,
		"paidByUserId":   id1,
		"participantIds": []uint{id1, id2},
	})

	code, _ := doJSON(t, app, "DELETE", "/api/v1/groups/"+path+"/members/"+gid(id3), nil)
	if code != 204 {
		t.Fatalf("expected 204 for unused member, status=%d", code)
	}
}

func TestDeleteGroup_blockedWhenMembersOrCharges(t *testing.T) {
	app := testApp(t)
	groupID, id1, id2, _ := seedTripWithMembers(t, app)
	path := gid(groupID)

	code, body := doJSON(t, app, "DELETE", "/api/v1/groups/"+path, nil)
	if code != 400 {
		t.Fatalf("expected 400 with members, status=%d body=%v", code, body)
	}
	if !strings.Contains(fmt.Sprint(body["error"]), "member") {
		t.Fatalf("expected members in error, body=%v", body)
	}

	// remove all members then add a charge-only blocker: need members to create charge
	doJSON(t, app, "POST", "/api/v1/groups/"+path+"/charges", map[string]any{
		"description":    "Snacks",
		"amount":         12,
		"paidByUserId":   id1,
		"participantIds": []uint{id1, id2},
	})
	code, body = doJSON(t, app, "DELETE", "/api/v1/groups/"+path, nil)
	if code != 400 {
		t.Fatalf("expected 400 with charges, status=%d body=%v", code, body)
	}
	if !strings.Contains(fmt.Sprint(body["error"]), "charge") {
		t.Fatalf("expected charges in error, body=%v", body)
	}
}

func TestDeleteGroup_allowedWhenEmpty(t *testing.T) {
	app := testApp(t)
	code, body := doJSON(t, app, "POST", "/api/v1/groups", map[string]any{
		"name":     "Empty group",
		"currency": "USD",
	})
	if code != 201 {
		t.Fatalf("create group: status=%d body=%v", code, body)
	}
	id := uint(responseData(body)["id"].(float64))
	code, _ = doJSON(t, app, "DELETE", "/api/v1/groups/"+gid(id), nil)
	if code != 204 {
		t.Fatalf("expected 204 for empty group, status=%d", code)
	}
}

func TestCategories_seededAndCRUD(t *testing.T) {
	app := testApp(t)
	code, body := doJSON(t, app, "GET", "/api/v1/categories", nil)
	if code != 200 {
		t.Fatalf("list categories: status=%d body=%v", code, body)
	}
	list, _ := body["data"].([]any)
	if len(list) < 10 {
		t.Fatalf("expected seeded builtins, got %d", len(list))
	}

	code, body = doJSON(t, app, "POST", "/api/v1/categories", map[string]any{
		"name": "Pets",
		"icon": "fa-solid fa-paw",
	})
	if code != 201 {
		t.Fatalf("create category: status=%d body=%v", code, body)
	}
	customID := uint(responseData(body)["id"].(float64))

	code, _ = doJSON(t, app, "PUT", "/api/v1/categories/"+gid(customID), map[string]any{
		"name": "Pets & Vet",
		"icon": "fa-solid fa-dog",
	})
	if code != 200 {
		t.Fatalf("update category: status=%d", code)
	}

	builtinID := uint(list[0].(map[string]any)["id"].(float64))
	code, body = doJSON(t, app, "PUT", "/api/v1/categories/"+gid(builtinID), map[string]any{
		"name": "Nope",
		"icon": "fa-solid fa-ban",
	})
	if code != 400 {
		t.Fatalf("expected 400 updating builtin, status=%d body=%v", code, body)
	}

	code, body = doJSON(t, app, "DELETE", "/api/v1/categories/"+gid(builtinID), nil)
	if code != 400 {
		t.Fatalf("expected 400 deleting builtin, status=%d body=%v", code, body)
	}

	code, _ = doJSON(t, app, "DELETE", "/api/v1/categories/"+gid(customID), nil)
	if code != 204 {
		t.Fatalf("expected 204 deleting custom, status=%d", code)
	}
}

func TestCharge_withCategoryAndSettleBreakdown(t *testing.T) {
	app := testApp(t)
	groupID, id1, id2, _ := seedTripWithMembers(t, app)
	path := gid(groupID)

	_, catsBody := doJSON(t, app, "GET", "/api/v1/categories", nil)
	cats := catsBody["data"].([]any)
	foodID := uint(0)
	for _, raw := range cats {
		c := raw.(map[string]any)
		if c["name"] == "Food" {
			foodID = uint(c["id"].(float64))
			break
		}
	}
	if foodID == 0 {
		t.Fatal("Food builtin missing")
	}

	code, body := doJSON(t, app, "POST", "/api/v1/groups/"+path+"/charges", map[string]any{
		"description":    "Dinner",
		"amount":         60,
		"paidByUserId":   id1,
		"categoryId":     foodID,
		"participantIds": []uint{id1, id2},
	})
	if code != 201 {
		t.Fatalf("create charge: status=%d body=%v", code, body)
	}
	data := responseData(body)
	if uint(data["categoryId"].(float64)) != foodID {
		t.Fatalf("categoryId missing on charge: %v", data)
	}
	if data["categoryName"] != "Food" {
		t.Fatalf("categoryName missing: %v", data)
	}
	if data["date"] == nil || data["date"] == "" {
		t.Fatalf("expected default date on charge: %v", data)
	}

	code, body = doJSON(t, app, "POST", "/api/v1/groups/"+path+"/settle", map[string]any{})
	if code != 200 {
		t.Fatalf("settle: status=%d body=%v", code, body)
	}
	settle := responseData(body)
	breakdown, _ := settle["categoryBreakdown"].([]any)
	if len(breakdown) == 0 {
		t.Fatalf("expected categoryBreakdown, got %v", settle)
	}
	row := breakdown[0].(map[string]any)
	if row["name"] != "Food" || row["amount"].(float64) != 60 {
		t.Fatalf("unexpected breakdown row: %v", row)
	}
	timeline, _ := settle["spendOverTime"].([]any)
	if len(timeline) == 0 {
		t.Fatalf("expected spendOverTime, got %v", settle)
	}
}
