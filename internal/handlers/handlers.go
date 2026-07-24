package handlers

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/uona-cmsc501-1-summer2026-nanlin/practicum-project/internal/models"
	"github.com/uona-cmsc501-1-summer2026-nanlin/practicum-project/internal/settle"
	"gorm.io/gorm"
)

// API holds the DB handle and Fiber route handlers.
type API struct {
	DB *gorm.DB
}

func (a *API) fail(c *fiber.Ctx, status int, code, msg string) error {
	return c.Status(status).JSON(models.ErrorBody{Error: msg, Code: code})
}

func parseID(c *fiber.Ctx, name string) (uint, error) {
	n, err := strconv.ParseUint(c.Params(name), 10, 64)
	return uint(n), err
}

func (a *API) loadGroup(id uint) (*models.Group, error) {
	var g models.Group
	err := a.DB.First(&g, id).Error
	return &g, err
}

func personInGroup(db *gorm.DB, groupID, personID uint) bool {
	var n int64
	db.Model(&models.Person{}).Where("id = ? AND group_id = ?", personID, groupID).Count(&n)
	return n > 0
}

func toChargeResponse(ch models.Charge) models.ChargeResponse {
	var ids []uint
	_ = json.Unmarshal([]byte(ch.ParticipantIDs), &ids)
	n := len(ids)
	share := 0.0
	if n > 0 {
		share = ch.Amount / float64(n)
		share = float64(int(share*100+0.5)) / 100
	}
	return models.ChargeResponse{
		ID:             ch.ID,
		GroupID:        ch.GroupID,
		Description:    ch.Description,
		Amount:         ch.Amount,
		PaidByPersonID: ch.PaidByPersonID,
		ParticipantIDs: ids,
		SplitRule:      ch.SplitRule,
		SharePerPerson: share,
	}
}

// --- Groups ---

func (a *API) CreateGroup(c *fiber.Ctx) error {
	var req models.CreateGroupRequest
	if err := c.BodyParser(&req); err != nil {
		return a.fail(c, 400, "BAD_REQUEST", "invalid JSON body")
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return a.fail(c, 400, "VALIDATION", "name is required")
	}
	if strings.TrimSpace(req.Currency) == "" {
		req.Currency = "USD"
	}
	g := models.Group{Name: req.Name, Currency: strings.ToUpper(req.Currency)}
	if err := a.DB.Create(&g).Error; err != nil {
		return a.fail(c, 500, "DB_ERROR", "could not create group")
	}
	return c.Status(201).JSON(g)
}

func (a *API) ListGroups(c *fiber.Ctx) error {
	q := a.DB.Model(&models.Group{})
	if name := strings.TrimSpace(c.Query("name")); name != "" {
		q = q.Where("name LIKE ?", "%"+name+"%")
	}
	var groups []models.Group
	if err := q.Order("id").Find(&groups).Error; err != nil {
		return a.fail(c, 500, "DB_ERROR", "could not list groups")
	}
	return c.JSON(groups)
}

func (a *API) GetGroup(c *fiber.Ctx) error {
	id, err := parseID(c, "groupId")
	if err != nil {
		return a.fail(c, 400, "VALIDATION", "invalid groupId")
	}
	g, err := a.loadGroup(id)
	if err != nil {
		return a.fail(c, 404, "NOT_FOUND", "group not found")
	}
	return c.JSON(g)
}

func (a *API) DeleteGroup(c *fiber.Ctx) error {
	id, err := parseID(c, "groupId")
	if err != nil {
		return a.fail(c, 400, "VALIDATION", "invalid groupId")
	}
	if _, err := a.loadGroup(id); err != nil {
		return a.fail(c, 404, "NOT_FOUND", "group not found")
	}
	a.DB.Where("group_id = ?", id).Delete(&models.Charge{})
	a.DB.Where("group_id = ?", id).Delete(&models.Person{})
	a.DB.Delete(&models.Group{}, id)
	return c.SendStatus(204)
}

// --- People ---

func (a *API) CreatePerson(c *fiber.Ctx) error {
	groupID, err := parseID(c, "groupId")
	if err != nil {
		return a.fail(c, 400, "VALIDATION", "invalid groupId")
	}
	if _, err := a.loadGroup(groupID); err != nil {
		return a.fail(c, 404, "NOT_FOUND", "group not found")
	}
	var req models.CreatePersonRequest
	if err := c.BodyParser(&req); err != nil {
		return a.fail(c, 400, "BAD_REQUEST", "invalid JSON body")
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return a.fail(c, 400, "VALIDATION", "name is required")
	}
	p := models.Person{GroupID: groupID, Name: req.Name, Email: strings.TrimSpace(req.Email)}
	if err := a.DB.Create(&p).Error; err != nil {
		return a.fail(c, 500, "DB_ERROR", "could not create person")
	}
	return c.Status(201).JSON(p)
}

func (a *API) ListPeople(c *fiber.Ctx) error {
	groupID, err := parseID(c, "groupId")
	if err != nil {
		return a.fail(c, 400, "VALIDATION", "invalid groupId")
	}
	if _, err := a.loadGroup(groupID); err != nil {
		return a.fail(c, 404, "NOT_FOUND", "group not found")
	}
	q := a.DB.Where("group_id = ?", groupID)
	if name := strings.TrimSpace(c.Query("name")); name != "" {
		q = q.Where("name LIKE ?", "%"+name+"%")
	}
	var people []models.Person
	if err := q.Order("id").Find(&people).Error; err != nil {
		return a.fail(c, 500, "DB_ERROR", "could not list people")
	}
	return c.JSON(people)
}

func (a *API) GetPerson(c *fiber.Ctx) error {
	groupID, err := parseID(c, "groupId")
	if err != nil {
		return a.fail(c, 400, "VALIDATION", "invalid groupId")
	}
	personID, err := parseID(c, "personId")
	if err != nil {
		return a.fail(c, 400, "VALIDATION", "invalid personId")
	}
	var p models.Person
	if err := a.DB.Where("id = ? AND group_id = ?", personID, groupID).First(&p).Error; err != nil {
		return a.fail(c, 404, "NOT_FOUND", "person not found in group")
	}
	return c.JSON(p)
}

func (a *API) DeletePerson(c *fiber.Ctx) error {
	groupID, err := parseID(c, "groupId")
	if err != nil {
		return a.fail(c, 400, "VALIDATION", "invalid groupId")
	}
	personID, err := parseID(c, "personId")
	if err != nil {
		return a.fail(c, 400, "VALIDATION", "invalid personId")
	}
	res := a.DB.Where("id = ? AND group_id = ?", personID, groupID).Delete(&models.Person{})
	if res.RowsAffected == 0 {
		return a.fail(c, 404, "NOT_FOUND", "person not found in group")
	}
	return c.SendStatus(204)
}

// --- Charges ---

func (a *API) CreateCharge(c *fiber.Ctx) error {
	groupID, err := parseID(c, "groupId")
	if err != nil {
		return a.fail(c, 400, "VALIDATION", "invalid groupId")
	}
	if _, err := a.loadGroup(groupID); err != nil {
		return a.fail(c, 404, "NOT_FOUND", "group not found")
	}
	var req models.CreateChargeRequest
	if err := c.BodyParser(&req); err != nil {
		return a.fail(c, 400, "BAD_REQUEST", "invalid JSON body")
	}
	req.Description = strings.TrimSpace(req.Description)
	if req.Description == "" {
		return a.fail(c, 400, "VALIDATION", "description is required")
	}
	if req.Amount <= 0 {
		return a.fail(c, 400, "VALIDATION", "amount must be greater than 0")
	}
	if req.PaidByPersonID == 0 || !personInGroup(a.DB, groupID, req.PaidByPersonID) {
		return a.fail(c, 404, "NOT_FOUND", "payer not found in group")
	}
	if strings.TrimSpace(req.SplitRule) == "" {
		req.SplitRule = "equal"
	}
	if !strings.EqualFold(req.SplitRule, "equal") {
		return a.fail(c, 400, "VALIDATION", "splitRule must be equal for MVP")
	}

	participantIDs := req.ParticipantIDs
	if len(participantIDs) == 0 {
		var people []models.Person
		a.DB.Where("group_id = ?", groupID).Find(&people)
		for _, p := range people {
			participantIDs = append(participantIDs, p.ID)
		}
	}
	if len(participantIDs) == 0 {
		return a.fail(c, 400, "VALIDATION", "group has no people to share the charge")
	}
	for _, pid := range participantIDs {
		if !personInGroup(a.DB, groupID, pid) {
			return a.fail(c, 404, "NOT_FOUND", "participant not found in group")
		}
	}

	raw, _ := json.Marshal(participantIDs)
	ch := models.Charge{
		GroupID:        groupID,
		Description:    req.Description,
		Amount:         req.Amount,
		PaidByPersonID: req.PaidByPersonID,
		ParticipantIDs: string(raw),
		SplitRule:      "equal",
	}
	if err := a.DB.Create(&ch).Error; err != nil {
		return a.fail(c, 500, "DB_ERROR", "could not create charge")
	}
	return c.Status(201).JSON(toChargeResponse(ch))
}

func (a *API) ListCharges(c *fiber.Ctx) error {
	groupID, err := parseID(c, "groupId")
	if err != nil {
		return a.fail(c, 400, "VALIDATION", "invalid groupId")
	}
	if _, err := a.loadGroup(groupID); err != nil {
		return a.fail(c, 404, "NOT_FOUND", "group not found")
	}
	q := a.DB.Where("group_id = ?", groupID)
	if paidBy := c.Query("paidBy"); paidBy != "" {
		id, err := strconv.ParseUint(paidBy, 10, 64)
		if err != nil {
			return a.fail(c, 400, "VALIDATION", "paidBy must be a number")
		}
		q = q.Where("paid_by_person_id = ?", id)
	}
	if min := c.Query("minAmount"); min != "" {
		v, err := strconv.ParseFloat(min, 64)
		if err != nil {
			return a.fail(c, 400, "VALIDATION", "minAmount must be a number")
		}
		q = q.Where("amount >= ?", v)
	}
	if max := c.Query("maxAmount"); max != "" {
		v, err := strconv.ParseFloat(max, 64)
		if err != nil {
			return a.fail(c, 400, "VALIDATION", "maxAmount must be a number")
		}
		q = q.Where("amount <= ?", v)
	}
	if text := strings.TrimSpace(c.Query("q")); text != "" {
		q = q.Where("description LIKE ?", "%"+text+"%")
	}
	var charges []models.Charge
	if err := q.Order("id").Find(&charges).Error; err != nil {
		return a.fail(c, 500, "DB_ERROR", "could not list charges")
	}
	out := make([]models.ChargeResponse, 0, len(charges))
	for _, ch := range charges {
		out = append(out, toChargeResponse(ch))
	}
	return c.JSON(out)
}

func (a *API) GetCharge(c *fiber.Ctx) error {
	groupID, err := parseID(c, "groupId")
	if err != nil {
		return a.fail(c, 400, "VALIDATION", "invalid groupId")
	}
	chargeID, err := parseID(c, "chargeId")
	if err != nil {
		return a.fail(c, 400, "VALIDATION", "invalid chargeId")
	}
	var ch models.Charge
	if err := a.DB.Where("id = ? AND group_id = ?", chargeID, groupID).First(&ch).Error; err != nil {
		return a.fail(c, 404, "NOT_FOUND", "charge not found in group")
	}
	return c.JSON(toChargeResponse(ch))
}

func (a *API) DeleteCharge(c *fiber.Ctx) error {
	groupID, err := parseID(c, "groupId")
	if err != nil {
		return a.fail(c, 400, "VALIDATION", "invalid groupId")
	}
	chargeID, err := parseID(c, "chargeId")
	if err != nil {
		return a.fail(c, 400, "VALIDATION", "invalid chargeId")
	}
	res := a.DB.Where("id = ? AND group_id = ?", chargeID, groupID).Delete(&models.Charge{})
	if res.RowsAffected == 0 {
		return a.fail(c, 404, "NOT_FOUND", "charge not found in group")
	}
	return c.SendStatus(204)
}

// --- Settle ---

func (a *API) Settle(c *fiber.Ctx) error {
	groupID, err := parseID(c, "groupId")
	if err != nil {
		return a.fail(c, 400, "VALIDATION", "invalid groupId")
	}
	g, err := a.loadGroup(groupID)
	if err != nil {
		return a.fail(c, 404, "NOT_FOUND", "group not found")
	}

	var req models.SettleRequest
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&req); err != nil {
			return a.fail(c, 400, "BAD_REQUEST", "invalid JSON body")
		}
	}

	var people []models.Person
	a.DB.Where("group_id = ?", groupID).Order("id").Find(&people)

	q := a.DB.Where("group_id = ?", groupID)
	if len(req.OnlyChargeIDs) > 0 {
		q = q.Where("id IN ?", req.OnlyChargeIDs)
	}
	var charges []models.Charge
	q.Order("id").Find(&charges)

	participantMap := make(map[uint][]uint, len(charges))
	for _, ch := range charges {
		var ids []uint
		_ = json.Unmarshal([]byte(ch.ParticipantIDs), &ids)
		participantMap[ch.ID] = ids
	}

	result := settle.Compute(*g, people, charges, participantMap)
	return c.JSON(result)
}

// Register mounts all /api/v1 routes on the Fiber app.
func (a *API) Register(app *fiber.App) {
	v1 := app.Group("/api/v1")

	v1.Post("/groups", a.CreateGroup)
	v1.Get("/groups", a.ListGroups)
	v1.Get("/groups/:groupId", a.GetGroup)
	v1.Delete("/groups/:groupId", a.DeleteGroup)

	v1.Post("/groups/:groupId/people", a.CreatePerson)
	v1.Get("/groups/:groupId/people", a.ListPeople)
	v1.Get("/groups/:groupId/people/:personId", a.GetPerson)
	v1.Delete("/groups/:groupId/people/:personId", a.DeletePerson)

	v1.Post("/groups/:groupId/charges", a.CreateCharge)
	v1.Get("/groups/:groupId/charges", a.ListCharges)
	v1.Get("/groups/:groupId/charges/:chargeId", a.GetCharge)
	v1.Delete("/groups/:groupId/charges/:chargeId", a.DeleteCharge)

	v1.Post("/groups/:groupId/settle", a.Settle)
}
