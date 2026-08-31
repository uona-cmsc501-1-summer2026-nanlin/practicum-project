package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/uona-cmsc501-1-summer2026-nanlin/practicum-project/internal/models"
	"gorm.io/gorm"
)

// TruncateResponse reports how many rows were removed per table.
type TruncateResponse struct {
	Charges int64 `json:"charges"`
	People  int64 `json:"people"`
	Groups  int64 `json:"groups"`
}

func (a *API) deleteAll(model any) (int64, error) {
	res := a.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(model)
	return res.RowsAffected, res.Error
}

func (a *API) resetAutoIncrement(tables ...string) {
	for _, table := range tables {
		a.DB.Exec("DELETE FROM sqlite_sequence WHERE name = ?", table)
	}
}

// TruncateCharges removes every charge row. Internal testing only.
func (a *API) TruncateCharges(c *fiber.Ctx) error {
	n, err := a.deleteAll(&models.Charge{})
	if err != nil {
		return errInternal("could not truncate charges")
	}
	a.resetAutoIncrement("charges")
	return respondOK(c, TruncateResponse{Charges: n})
}

// TruncatePeople removes every person row. Internal testing only.
func (a *API) TruncatePeople(c *fiber.Ctx) error {
	n, err := a.deleteAll(&models.Person{})
	if err != nil {
		return errInternal("could not truncate people")
	}
	a.resetAutoIncrement("people")
	return respondOK(c, TruncateResponse{People: n})
}

// TruncateGroups removes all groups and their people and charges. Internal testing only.
func (a *API) TruncateGroups(c *fiber.Ctx) error {
	out, err := a.truncateAllData()
	if err != nil {
		return errInternal("could not truncate groups")
	}
	return respondOK(c, out)
}

// TruncateAll removes every charge, person, and group. Internal testing only.
func (a *API) TruncateAll(c *fiber.Ctx) error {
	out, err := a.truncateAllData()
	if err != nil {
		return errInternal("could not truncate all data")
	}
	return respondOK(c, out)
}

func (a *API) truncateAllData() (TruncateResponse, error) {
	charges, err := a.deleteAll(&models.Charge{})
	if err != nil {
		return TruncateResponse{}, err
	}
	people, err := a.deleteAll(&models.Person{})
	if err != nil {
		return TruncateResponse{}, err
	}
	groups, err := a.deleteAll(&models.Group{})
	if err != nil {
		return TruncateResponse{}, err
	}
	a.resetAutoIncrement("charges", "people", "groups")
	return TruncateResponse{Charges: charges, People: people, Groups: groups}, nil
}
