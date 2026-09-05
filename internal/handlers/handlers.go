package handlers

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/uona-cmsc501-1-summer2026-nanlin/practicum-project/internal/models"
	"github.com/uona-cmsc501-1-summer2026-nanlin/practicum-project/internal/settle"
	"gorm.io/gorm"
)

// API holds the DB handle and Fiber route handlers.
type API struct {
	DB *gorm.DB
}

func parseID(c *fiber.Ctx, name string) (uint, error) {
	n, err := strconv.ParseUint(c.Params(name), 10, 64)
	return uint(n), err
}

func plural(n int64) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func (a *API) loadGroup(id uint) (*models.Group, error) {
	var g models.Group
	err := a.DB.First(&g, id).Error
	return &g, err
}

func (a *API) loadUser(id uint) (*models.User, error) {
	var u models.User
	err := a.DB.First(&u, id).Error
	return &u, err
}

func userInGroup(db *gorm.DB, groupID, userID uint) bool {
	var n int64
	db.Model(&models.GroupMember{}).Where("group_id = ? AND user_id = ?", groupID, userID).Count(&n)
	return n > 0
}

func (a *API) membersAsUsers(groupID uint) ([]models.User, error) {
	var members []models.GroupMember
	if err := a.DB.Where("group_id = ?", groupID).Order("user_id").Find(&members).Error; err != nil {
		return nil, err
	}
	if len(members) == 0 {
		return []models.User{}, nil
	}
	ids := make([]uint, len(members))
	for i, m := range members {
		ids[i] = m.UserID
	}
	var users []models.User
	if err := a.DB.Where("id IN ?", ids).Order("id").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func chargeReferencesUser(db *gorm.DB, userID uint) bool {
	var n int64
	db.Model(&models.Charge{}).Where("paid_by_user_id = ?", userID).Count(&n)
	if n > 0 {
		return true
	}
	var charges []models.Charge
	db.Find(&charges)
	for _, ch := range charges {
		var ids []uint
		_ = json.Unmarshal([]byte(ch.ParticipantIDs), &ids)
		for _, id := range ids {
			if id == userID {
				return true
			}
		}
	}
	return false
}

// groupChargesForUser returns group charges where the user is payer or participant.
func groupChargesForUser(db *gorm.DB, groupID, userID uint) []models.Charge {
	var charges []models.Charge
	db.Where("group_id = ?", groupID).Order("id").Find(&charges)
	out := make([]models.Charge, 0)
	for _, ch := range charges {
		if ch.PaidByUserID == userID {
			out = append(out, ch)
			continue
		}
		var ids []uint
		_ = json.Unmarshal([]byte(ch.ParticipantIDs), &ids)
		for _, id := range ids {
			if id == userID {
				out = append(out, ch)
				break
			}
		}
	}
	return out
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
		Date:           ch.Date,
		PaidByUserID:   ch.PaidByUserID,
		CategoryID:     ch.CategoryID,
		ParticipantIDs: ids,
		SplitRule:      ch.SplitRule,
		SharePerPerson: share,
	}
}

func (a *API) enrichChargeResponse(resp models.ChargeResponse) models.ChargeResponse {
	if resp.CategoryID == nil {
		return resp
	}
	var cat models.Category
	if err := a.DB.First(&cat, *resp.CategoryID).Error; err != nil {
		return resp
	}
	resp.CategoryName = cat.Name
	resp.CategoryIcon = cat.Icon
	return resp
}

func (a *API) loadCategory(id uint) (*models.Category, error) {
	var cat models.Category
	err := a.DB.First(&cat, id).Error
	return &cat, err
}

func (a *API) validateCategoryID(id *uint) error {
	if id == nil {
		return nil
	}
	if *id == 0 {
		return errValidation("categoryId must be greater than 0")
	}
	if _, err := a.loadCategory(*id); err != nil {
		return errNotFound("category not found")
	}
	return nil
}

func (a *API) categoryBreakdown(charges []models.Charge) []models.CategorySpendRow {
	type agg struct {
		amount float64
		count  int
		id     *uint
	}
	byKey := map[string]*agg{}
	order := []string{}
	for _, ch := range charges {
		key := "uncategorized"
		var id *uint
		if ch.CategoryID != nil {
			key = fmt.Sprintf("c:%d", *ch.CategoryID)
			cid := *ch.CategoryID
			id = &cid
		}
		if byKey[key] == nil {
			byKey[key] = &agg{id: id}
			order = append(order, key)
		}
		byKey[key].amount += ch.Amount
		byKey[key].count++
	}

	catCache := map[uint]models.Category{}
	out := make([]models.CategorySpendRow, 0, len(order))
	for _, key := range order {
		aRow := byKey[key]
		row := models.CategorySpendRow{
			CategoryID:  aRow.id,
			Name:        "Uncategorized",
			Icon:        "fa-solid fa-circle-question",
			Amount:      float64(int(aRow.amount*100+0.5)) / 100,
			ChargeCount: aRow.count,
		}
		if aRow.id != nil {
			cat, ok := catCache[*aRow.id]
			if !ok {
				if c, err := a.loadCategory(*aRow.id); err == nil {
					cat = *c
					catCache[*aRow.id] = cat
				}
			}
			if cat.ID != 0 {
				row.Name = cat.Name
				row.Icon = cat.Icon
			} else {
				row.Name = "Unknown category"
			}
		}
		out = append(out, row)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Amount == out[j].Amount {
			return out[i].Name < out[j].Name
		}
		return out[i].Amount > out[j].Amount
	})
	return out
}

func parseChargeDate(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return time.Now().Format("2006-01-02"), nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return "", errValidation("date must be YYYY-MM-DD")
	}
	return t.Format("2006-01-02"), nil
}

func (a *API) spendOverTime(charges []models.Charge) []models.SpendDayRow {
	type agg struct {
		amount float64
		count  int
	}
	byDate := map[string]*agg{}
	for _, ch := range charges {
		d := strings.TrimSpace(ch.Date)
		if d == "" {
			d = "unknown"
		}
		if byDate[d] == nil {
			byDate[d] = &agg{}
		}
		byDate[d].amount += ch.Amount
		byDate[d].count++
	}
	out := make([]models.SpendDayRow, 0, len(byDate))
	for d, aRow := range byDate {
		out = append(out, models.SpendDayRow{
			Date:        d,
			Amount:      float64(int(aRow.amount*100+0.5)) / 100,
			ChargeCount: aRow.count,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Date < out[j].Date
	})
	return out
}

// --- Groups ---

func (a *API) CreateGroup(c *fiber.Ctx) error {
	var req models.CreateGroupRequest
	if err := c.BodyParser(&req); err != nil {
		return errBadRequest("invalid JSON body")
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return errValidation("name is required")
	}
	currency := models.NormalizeCurrency(req.Currency)
	if !models.IsAllowedCurrency(currency) {
		return errValidation("currency must be one of: " + strings.Join(models.CurrencyCodes(), ", "))
	}
	g := models.Group{Name: req.Name, Currency: currency}
	if err := a.DB.Create(&g).Error; err != nil {
		return errInternal("could not create group")
	}
	return respondCreated(c, g)
}

func (a *API) ListCurrencies(c *fiber.Ctx) error {
	return respondOK(c, models.Currencies)
}

func (a *API) ListGroups(c *fiber.Ctx) error {
	q := a.DB.Model(&models.Group{})
	if name := strings.TrimSpace(c.Query("name")); name != "" {
		q = q.Where("name LIKE ?", "%"+name+"%")
	}
	var groups []models.Group
	if err := q.Order("id").Find(&groups).Error; err != nil {
		return errInternal("could not list groups")
	}
	return respondOK(c, groups)
}

func (a *API) GetGroup(c *fiber.Ctx) error {
	id, err := parseID(c, "groupId")
	if err != nil {
		return errValidation("invalid groupId")
	}
	g, err := a.loadGroup(id)
	if err != nil {
		return errNotFound("group not found")
	}
	return respondOK(c, g)
}

func (a *API) UpdateGroup(c *fiber.Ctx) error {
	id, err := parseID(c, "groupId")
	if err != nil {
		return errValidation("invalid groupId")
	}
	g, err := a.loadGroup(id)
	if err != nil {
		return errNotFound("group not found")
	}
	var req models.UpdateGroupRequest
	if err := c.BodyParser(&req); err != nil {
		return errBadRequest("invalid JSON body")
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return errValidation("name is required")
	}
	currency := models.NormalizeCurrency(req.Currency)
	if !models.IsAllowedCurrency(currency) {
		return errValidation("currency must be one of: " + strings.Join(models.CurrencyCodes(), ", "))
	}
	g.Name = req.Name
	g.Currency = currency
	if err := a.DB.Save(g).Error; err != nil {
		return errInternal("could not update group")
	}
	return respondOK(c, g)
}

func (a *API) DeleteGroup(c *fiber.Ctx) error {
	id, err := parseID(c, "groupId")
	if err != nil {
		return errValidation("invalid groupId")
	}
	if _, err := a.loadGroup(id); err != nil {
		return errNotFound("group not found")
	}
	var memberCount, chargeCount int64
	a.DB.Model(&models.GroupMember{}).Where("group_id = ?", id).Count(&memberCount)
	a.DB.Model(&models.Charge{}).Where("group_id = ?", id).Count(&chargeCount)
	if memberCount > 0 || chargeCount > 0 {
		parts := make([]string, 0, 2)
		if memberCount > 0 {
			parts = append(parts, fmt.Sprintf("%d member%s", memberCount, plural(memberCount)))
		}
		if chargeCount > 0 {
			parts = append(parts, fmt.Sprintf("%d charge%s", chargeCount, plural(chargeCount)))
		}
		return errValidation("cannot delete group while it still has " + strings.Join(parts, " and "))
	}
	a.DB.Delete(&models.Group{}, id)
	return respondNoContent(c)
}

// --- People (global users) ---

func (a *API) CreatePerson(c *fiber.Ctx) error {
	var req models.CreatePersonRequest
	if err := c.BodyParser(&req); err != nil {
		return errBadRequest("invalid JSON body")
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return errValidation("name is required")
	}
	u := models.User{Name: req.Name, Email: strings.TrimSpace(req.Email)}
	if err := a.DB.Create(&u).Error; err != nil {
		return errInternal("could not create person")
	}
	return respondCreated(c, u)
}

func (a *API) ListPeople(c *fiber.Ctx) error {
	q := a.DB.Model(&models.User{})
	if name := strings.TrimSpace(c.Query("name")); name != "" {
		q = q.Where("name LIKE ?", "%"+name+"%")
	}
	var users []models.User
	if err := q.Order("id").Find(&users).Error; err != nil {
		return errInternal("could not list people")
	}
	return respondOK(c, users)
}

func (a *API) GetPerson(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return errValidation("invalid id")
	}
	u, err := a.loadUser(id)
	if err != nil {
		return errNotFound("person not found")
	}
	return respondOK(c, u)
}

func (a *API) UpdatePerson(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return errValidation("invalid id")
	}
	u, err := a.loadUser(id)
	if err != nil {
		return errNotFound("person not found")
	}
	var req models.UpdatePersonRequest
	if err := c.BodyParser(&req); err != nil {
		return errBadRequest("invalid JSON body")
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return errValidation("name is required")
	}
	u.Name = req.Name
	u.Email = strings.TrimSpace(req.Email)
	if err := a.DB.Save(u).Error; err != nil {
		return errInternal("could not update person")
	}
	return respondOK(c, u)
}

func (a *API) DeletePerson(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return errValidation("invalid id")
	}
	if _, err := a.loadUser(id); err != nil {
		return errNotFound("person not found")
	}
	var memberCount int64
	a.DB.Model(&models.GroupMember{}).Where("user_id = ?", id).Count(&memberCount)
	if memberCount > 0 {
		return errValidation("cannot delete person who is still in a group")
	}
	if chargeReferencesUser(a.DB, id) {
		return errValidation("cannot delete person referenced by a charge")
	}
	a.DB.Delete(&models.User{}, id)
	return respondNoContent(c)
}

// --- Categories (global) ---

func (a *API) ListCategories(c *fiber.Ctx) error {
	var cats []models.Category
	if err := a.DB.Order("builtin DESC, id").Find(&cats).Error; err != nil {
		return errInternal("could not list categories")
	}
	return respondOK(c, cats)
}

func (a *API) CreateCategory(c *fiber.Ctx) error {
	var req models.CreateCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return errBadRequest("invalid JSON body")
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Icon = strings.TrimSpace(req.Icon)
	if req.Name == "" {
		return errValidation("name is required")
	}
	if req.Icon == "" {
		return errValidation("icon is required")
	}
	cat := models.Category{Name: req.Name, Icon: req.Icon, Builtin: false}
	if err := a.DB.Create(&cat).Error; err != nil {
		return errInternal("could not create category")
	}
	return respondCreated(c, cat)
}

func (a *API) UpdateCategory(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return errValidation("invalid id")
	}
	cat, err := a.loadCategory(id)
	if err != nil {
		return errNotFound("category not found")
	}
	if cat.Builtin {
		return errValidation("cannot edit a built-in category")
	}
	var req models.UpdateCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return errBadRequest("invalid JSON body")
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Icon = strings.TrimSpace(req.Icon)
	if req.Name == "" {
		return errValidation("name is required")
	}
	if req.Icon == "" {
		return errValidation("icon is required")
	}
	cat.Name = req.Name
	cat.Icon = req.Icon
	if err := a.DB.Save(cat).Error; err != nil {
		return errInternal("could not update category")
	}
	return respondOK(c, cat)
}

func (a *API) DeleteCategory(c *fiber.Ctx) error {
	id, err := parseID(c, "id")
	if err != nil {
		return errValidation("invalid id")
	}
	cat, err := a.loadCategory(id)
	if err != nil {
		return errNotFound("category not found")
	}
	if cat.Builtin {
		return errValidation("cannot delete a built-in category")
	}
	var n int64
	a.DB.Model(&models.Charge{}).Where("category_id = ?", id).Count(&n)
	if n > 0 {
		return errValidation(fmt.Sprintf("cannot delete category used by %d charge%s", n, plural(n)))
	}
	a.DB.Delete(&models.Category{}, id)
	return respondNoContent(c)
}

// --- Members ---

func (a *API) ListMembers(c *fiber.Ctx) error {
	groupID, err := parseID(c, "groupId")
	if err != nil {
		return errValidation("invalid groupId")
	}
	if _, err := a.loadGroup(groupID); err != nil {
		return errNotFound("group not found")
	}
	users, err := a.membersAsUsers(groupID)
	if err != nil {
		return errInternal("could not list members")
	}
	return respondOK(c, users)
}

func (a *API) AddMember(c *fiber.Ctx) error {
	groupID, err := parseID(c, "groupId")
	if err != nil {
		return errValidation("invalid groupId")
	}
	if _, err := a.loadGroup(groupID); err != nil {
		return errNotFound("group not found")
	}
	var req models.AddMemberRequest
	if err := c.BodyParser(&req); err != nil {
		return errBadRequest("invalid JSON body")
	}
	if req.UserID == 0 {
		return errValidation("userId is required")
	}
	u, err := a.loadUser(req.UserID)
	if err != nil {
		return errNotFound("person not found")
	}
	if userInGroup(a.DB, groupID, req.UserID) {
		return errValidation("person is already a member of this group")
	}
	m := models.GroupMember{GroupID: groupID, UserID: req.UserID}
	if err := a.DB.Create(&m).Error; err != nil {
		return errInternal("could not add member")
	}
	return respondCreated(c, u)
}

func (a *API) RemoveMember(c *fiber.Ctx) error {
	groupID, err := parseID(c, "groupId")
	if err != nil {
		return errValidation("invalid groupId")
	}
	userID, err := parseID(c, "userId")
	if err != nil {
		return errValidation("invalid userId")
	}
	if _, err := a.loadGroup(groupID); err != nil {
		return errNotFound("group not found")
	}
	if !userInGroup(a.DB, groupID, userID) {
		return errNotFound("member not found in group")
	}
	blocking := groupChargesForUser(a.DB, groupID, userID)
	if len(blocking) > 0 {
		names := make([]string, len(blocking))
		for i, ch := range blocking {
			names[i] = ch.Description
		}
		return errValidation("cannot remove member who is part of charges: " + strings.Join(names, ", "))
	}
	res := a.DB.Where("group_id = ? AND user_id = ?", groupID, userID).Delete(&models.GroupMember{})
	if res.RowsAffected == 0 {
		return errNotFound("member not found in group")
	}
	return respondNoContent(c)
}

// --- Charges ---

func (a *API) CreateCharge(c *fiber.Ctx) error {
	groupID, err := parseID(c, "groupId")
	if err != nil {
		return errValidation("invalid groupId")
	}
	if _, err := a.loadGroup(groupID); err != nil {
		return errNotFound("group not found")
	}
	var req models.CreateChargeRequest
	if err := c.BodyParser(&req); err != nil {
		return errBadRequest("invalid JSON body")
	}
	req.Description = strings.TrimSpace(req.Description)
	if req.Description == "" {
		return errValidation("description is required")
	}
	if req.Amount <= 0 {
		return errValidation("amount must be greater than 0")
	}
	if req.PaidByUserID == 0 || !userInGroup(a.DB, groupID, req.PaidByUserID) {
		return errNotFound("payer not found in group")
	}
	if strings.TrimSpace(req.SplitRule) == "" {
		req.SplitRule = "equal"
	}
	if !strings.EqualFold(req.SplitRule, "equal") {
		return errValidation("splitRule must be equal for MVP")
	}

	if err := a.validateCategoryID(req.CategoryID); err != nil {
		return err
	}
	chargeDate, err := parseChargeDate(req.Date)
	if err != nil {
		return err
	}

	participantIDs := req.ParticipantIDs
	if len(participantIDs) == 0 {
		users, err := a.membersAsUsers(groupID)
		if err != nil {
			return errInternal("could not load members")
		}
		for _, u := range users {
			participantIDs = append(participantIDs, u.ID)
		}
	}
	if len(participantIDs) == 0 {
		return errValidation("group has no members to share the charge")
	}
	for _, uid := range participantIDs {
		if !userInGroup(a.DB, groupID, uid) {
			return errNotFound("participant not found in group")
		}
	}

	raw, _ := json.Marshal(participantIDs)
	ch := models.Charge{
		GroupID:        groupID,
		Description:    req.Description,
		Amount:         req.Amount,
		Date:           chargeDate,
		PaidByUserID:   req.PaidByUserID,
		CategoryID:     req.CategoryID,
		ParticipantIDs: string(raw),
		SplitRule:      "equal",
	}
	if err := a.DB.Create(&ch).Error; err != nil {
		return errInternal("could not create charge")
	}
	return respondCreated(c, a.enrichChargeResponse(toChargeResponse(ch)))
}

func (a *API) ListCharges(c *fiber.Ctx) error {
	groupID, err := parseID(c, "groupId")
	if err != nil {
		return errValidation("invalid groupId")
	}
	if _, err := a.loadGroup(groupID); err != nil {
		return errNotFound("group not found")
	}
	q := a.DB.Where("group_id = ?", groupID)
	if paidBy := c.Query("paidBy"); paidBy != "" {
		id, err := strconv.ParseUint(paidBy, 10, 64)
		if err != nil {
			return errValidation("paidBy must be a number")
		}
		q = q.Where("paid_by_user_id = ?", id)
	}
	if min := c.Query("minAmount"); min != "" {
		v, err := strconv.ParseFloat(min, 64)
		if err != nil {
			return errValidation("minAmount must be a number")
		}
		q = q.Where("amount >= ?", v)
	}
	if max := c.Query("maxAmount"); max != "" {
		v, err := strconv.ParseFloat(max, 64)
		if err != nil {
			return errValidation("maxAmount must be a number")
		}
		q = q.Where("amount <= ?", v)
	}
	if text := strings.TrimSpace(c.Query("q")); text != "" {
		q = q.Where("description LIKE ?", "%"+text+"%")
	}
	var charges []models.Charge
	if err := q.Order("id").Find(&charges).Error; err != nil {
		return errInternal("could not list charges")
	}
	out := make([]models.ChargeResponse, 0, len(charges))
	for _, ch := range charges {
		out = append(out, a.enrichChargeResponse(toChargeResponse(ch)))
	}
	return respondOK(c, out)
}

func (a *API) GetCharge(c *fiber.Ctx) error {
	groupID, err := parseID(c, "groupId")
	if err != nil {
		return errValidation("invalid groupId")
	}
	chargeID, err := parseID(c, "chargeId")
	if err != nil {
		return errValidation("invalid chargeId")
	}
	var ch models.Charge
	if err := a.DB.Where("id = ? AND group_id = ?", chargeID, groupID).First(&ch).Error; err != nil {
		return errNotFound("charge not found in group")
	}
	return respondOK(c, a.enrichChargeResponse(toChargeResponse(ch)))
}

func (a *API) UpdateCharge(c *fiber.Ctx) error {
	groupID, err := parseID(c, "groupId")
	if err != nil {
		return errValidation("invalid groupId")
	}
	chargeID, err := parseID(c, "chargeId")
	if err != nil {
		return errValidation("invalid chargeId")
	}
	var ch models.Charge
	if err := a.DB.Where("id = ? AND group_id = ?", chargeID, groupID).First(&ch).Error; err != nil {
		return errNotFound("charge not found in group")
	}
	var req models.UpdateChargeRequest
	if err := c.BodyParser(&req); err != nil {
		return errBadRequest("invalid JSON body")
	}
	req.Description = strings.TrimSpace(req.Description)
	if req.Description == "" {
		return errValidation("description is required")
	}
	if req.Amount <= 0 {
		return errValidation("amount must be greater than 0")
	}
	if req.PaidByUserID == 0 || !userInGroup(a.DB, groupID, req.PaidByUserID) {
		return errNotFound("payer not found in group")
	}
	if strings.TrimSpace(req.SplitRule) == "" {
		req.SplitRule = "equal"
	}
	if !strings.EqualFold(req.SplitRule, "equal") {
		return errValidation("splitRule must be equal for MVP")
	}
	if err := a.validateCategoryID(req.CategoryID); err != nil {
		return err
	}
	chargeDate, err := parseChargeDate(req.Date)
	if err != nil {
		return err
	}
	participantIDs := req.ParticipantIDs
	if len(participantIDs) == 0 {
		return errValidation("at least one participant is required")
	}
	for _, uid := range participantIDs {
		if !userInGroup(a.DB, groupID, uid) {
			return errNotFound("participant not found in group")
		}
	}
	raw, _ := json.Marshal(participantIDs)
	ch.Description = req.Description
	ch.Amount = req.Amount
	ch.Date = chargeDate
	ch.PaidByUserID = req.PaidByUserID
	ch.CategoryID = req.CategoryID
	ch.ParticipantIDs = string(raw)
	ch.SplitRule = "equal"
	if err := a.DB.Save(&ch).Error; err != nil {
		return errInternal("could not update charge")
	}
	return respondOK(c, a.enrichChargeResponse(toChargeResponse(ch)))
}

func (a *API) DeleteCharge(c *fiber.Ctx) error {
	groupID, err := parseID(c, "groupId")
	if err != nil {
		return errValidation("invalid groupId")
	}
	chargeID, err := parseID(c, "chargeId")
	if err != nil {
		return errValidation("invalid chargeId")
	}
	res := a.DB.Where("id = ? AND group_id = ?", chargeID, groupID).Delete(&models.Charge{})
	if res.RowsAffected == 0 {
		return errNotFound("charge not found in group")
	}
	return respondNoContent(c)
}

// --- Settle ---

func (a *API) Settle(c *fiber.Ctx) error {
	groupID, err := parseID(c, "groupId")
	if err != nil {
		return errValidation("invalid groupId")
	}
	g, err := a.loadGroup(groupID)
	if err != nil {
		return errNotFound("group not found")
	}

	var req models.SettleRequest
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&req); err != nil {
			return errBadRequest("invalid JSON body")
		}
	}

	users, err := a.membersAsUsers(groupID)
	if err != nil {
		return errInternal("could not load members")
	}

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

	result := settle.Compute(*g, users, charges, participantMap)
	result.CategoryBreakdown = a.categoryBreakdown(charges)
	result.SpendOverTime = a.spendOverTime(charges)
	return respondOK(c, result)
}

// Register mounts all /api/v1 routes on the Fiber app.
func (a *API) Register(app *fiber.App) {
	v1 := app.Group("/api/v1")

	v1.Get("/currencies", a.ListCurrencies)

	v1.Post("/groups", a.CreateGroup)
	v1.Get("/groups", a.ListGroups)
	v1.Get("/groups/:groupId", a.GetGroup)
	v1.Put("/groups/:groupId", a.UpdateGroup)
	v1.Delete("/groups/:groupId", a.DeleteGroup)

	v1.Post("/people", a.CreatePerson)
	v1.Get("/people", a.ListPeople)
	v1.Get("/people/:id", a.GetPerson)
	v1.Put("/people/:id", a.UpdatePerson)
	v1.Delete("/people/:id", a.DeletePerson)

	v1.Get("/categories", a.ListCategories)
	v1.Post("/categories", a.CreateCategory)
	v1.Put("/categories/:id", a.UpdateCategory)
	v1.Delete("/categories/:id", a.DeleteCategory)

	v1.Get("/groups/:groupId/members", a.ListMembers)
	v1.Post("/groups/:groupId/members", a.AddMember)
	v1.Delete("/groups/:groupId/members/:userId", a.RemoveMember)

	v1.Post("/groups/:groupId/charges", a.CreateCharge)
	v1.Get("/groups/:groupId/charges", a.ListCharges)
	v1.Get("/groups/:groupId/charges/:chargeId", a.GetCharge)
	v1.Put("/groups/:groupId/charges/:chargeId", a.UpdateCharge)
	v1.Delete("/groups/:groupId/charges/:chargeId", a.DeleteCharge)

	v1.Post("/groups/:groupId/settle", a.Settle)

	internal := v1.Group("/internal/testing")
	internal.Delete("/truncate/charges", a.TruncateCharges)
	internal.Delete("/truncate/people", a.TruncatePeople)
	internal.Delete("/truncate/groups", a.TruncateGroups)
	internal.Delete("/truncate/all", a.TruncateAll)
}
