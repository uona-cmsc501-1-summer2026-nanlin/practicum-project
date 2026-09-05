package settle

import (
	"testing"

	"github.com/uona-cmsc501-1-summer2026-nanlin/practicum-project/internal/models"
)

func TestCompute_equalSplitDinner(t *testing.T) {
	group := models.Group{ID: 1, Currency: "USD"}
	users := []models.User{
		{ID: 1, Name: "Alex"},
		{ID: 2, Name: "Sam"},
		{ID: 3, Name: "Jordan"},
	}
	charges := []models.Charge{
		{ID: 1, PaidByUserID: 1, Amount: 90},
	}
	participants := map[uint][]uint{1: {1, 2, 3}}

	result := Compute(group, users, charges, participants)

	if len(result.Balances) != 3 {
		t.Fatalf("expected 3 balances, got %d", len(result.Balances))
	}
	if result.Balances[0].Net != 60 || result.Balances[1].Net != -30 || result.Balances[2].Net != -30 {
		t.Fatalf("unexpected balances: %+v", result.Balances)
	}
	if len(result.Transfers) != 2 {
		t.Fatalf("expected 2 transfers, got %d", len(result.Transfers))
	}
	if result.Transfers[0].Amount != 30 || result.Transfers[1].Amount != 30 {
		t.Fatalf("unexpected transfers: %+v", result.Transfers)
	}
}

func TestCompute_noCharges(t *testing.T) {
	group := models.Group{ID: 1, Currency: "USD"}
	users := []models.User{{ID: 1, Name: "Alex"}}
	result := Compute(group, users, nil, nil)

	if len(result.Balances) != 1 || result.Balances[0].Net != 0 {
		t.Fatalf("expected zero balance, got %+v", result.Balances)
	}
	if len(result.Transfers) != 0 {
		t.Fatalf("expected no transfers, got %+v", result.Transfers)
	}
}

func TestCompute_twoCharges(t *testing.T) {
	group := models.Group{ID: 1, Currency: "USD"}
	users := []models.User{
		{ID: 1, Name: "Alex"},
		{ID: 2, Name: "Sam"},
	}
	charges := []models.Charge{
		{ID: 1, PaidByUserID: 1, Amount: 40},
		{ID: 2, PaidByUserID: 2, Amount: 20},
	}
	participants := map[uint][]uint{
		1: {1, 2},
		2: {1, 2},
	}

	result := Compute(group, users, charges, participants)

	// Alex paid 40, owes 20+10=30 -> +10; Sam paid 20, owes 20+10=30 -> -10
	if result.Balances[0].Net != 10 || result.Balances[1].Net != -10 {
		t.Fatalf("unexpected balances: %+v", result.Balances)
	}
	if len(result.Transfers) != 1 || result.Transfers[0].Amount != 10 {
		t.Fatalf("unexpected transfers: %+v", result.Transfers)
	}
}
