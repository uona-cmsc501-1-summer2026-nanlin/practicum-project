package settle

import (
	"math"
	"sort"

	"github.com/uona-cmsc501-1-summer2026-nanlin/practicum-project/internal/models"
)

const epsilon = 0.005

// Compute builds net balances and simplified transfers from people and charges.
// Each charge: payer gets +amount; each participant owes amount/n (equal split).
func Compute(group models.Group, people []models.Person, charges []models.Charge, participantMap map[uint][]uint) models.SettleResponse {
	balance := make(map[uint]float64, len(people))
	names := make(map[uint]string, len(people))
	for _, p := range people {
		balance[p.ID] = 0
		names[p.ID] = p.Name
	}

	for _, c := range charges {
		ids := participantMap[c.ID]
		if len(ids) == 0 {
			continue
		}
		balance[c.PaidByPersonID] += c.Amount
		share := c.Amount / float64(len(ids))
		for _, pid := range ids {
			balance[pid] -= share
		}
	}

	balances := make([]models.BalanceRow, 0, len(people))
	for _, p := range people {
		net := round2(balance[p.ID])
		balances = append(balances, models.BalanceRow{
			PersonID: p.ID,
			Name:     p.Name,
			Net:      net,
		})
	}
	sort.Slice(balances, func(i, j int) bool { return balances[i].PersonID < balances[j].PersonID })

	type party struct {
		id  uint
		amt float64
	}
	var creditors, debtors []party
	for id, net := range balance {
		n := round2(net)
		if n > epsilon {
			creditors = append(creditors, party{id, n})
		} else if n < -epsilon {
			debtors = append(debtors, party{id, -n})
		}
	}
	sort.Slice(creditors, func(i, j int) bool { return creditors[i].id < creditors[j].id })
	sort.Slice(debtors, func(i, j int) bool { return debtors[i].id < debtors[j].id })

	transfers := make([]models.TransferRow, 0)
	i, j := 0, 0
	for i < len(debtors) && j < len(creditors) {
		pay := math.Min(debtors[i].amt, creditors[j].amt)
		pay = round2(pay)
		if pay > epsilon {
			transfers = append(transfers, models.TransferRow{
				FromPersonID: debtors[i].id,
				FromName:     names[debtors[i].id],
				ToPersonID:   creditors[j].id,
				ToName:       names[creditors[j].id],
				Amount:       pay,
			})
		}
		debtors[i].amt = round2(debtors[i].amt - pay)
		creditors[j].amt = round2(creditors[j].amt - pay)
		if debtors[i].amt <= epsilon {
			i++
		}
		if creditors[j].amt <= epsilon {
			j++
		}
	}

	return models.SettleResponse{
		GroupID:   group.ID,
		Currency:  group.Currency,
		Balances:  balances,
		Transfers: transfers,
	}
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
