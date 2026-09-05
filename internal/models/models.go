package models

import "time"

// Group is one split session (trip, apartment month, etc.).
type Group struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"not null"`
	Currency  string    `json:"currency" gorm:"not null;default:USD"`
	CreatedAt time.Time `json:"createdAt"`
}

// User is a global person identity (not tied to one group).
type User struct {
	ID    uint   `json:"id" gorm:"primaryKey"`
	Name  string `json:"name" gorm:"not null"`
	Email string `json:"email"`
}

// GroupMember links a user to a group (many-to-many).
type GroupMember struct {
	GroupID uint `json:"groupId" gorm:"primaryKey"`
	UserID  uint `json:"userId" gorm:"primaryKey"`
}

// Category is a global expense category (builtin or user-defined).
type Category struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"not null"`
	Icon      string    `json:"icon" gorm:"not null"` // e.g. "fa-solid fa-utensils"
	Builtin   bool      `json:"builtin" gorm:"not null;default:false"`
	CreatedAt time.Time `json:"createdAt"`
}

// Charge is a shared expense within a group.
// ParticipantIDs is stored as JSON text (user IDs who share this charge).
type Charge struct {
	ID             uint    `json:"id" gorm:"primaryKey"`
	GroupID        uint    `json:"groupId" gorm:"index;not null"`
	Description    string  `json:"description" gorm:"not null"`
	Amount         float64 `json:"amount" gorm:"not null"`
	Date           string  `json:"date" gorm:"not null;default:'';index;size:10"` // YYYY-MM-DD
	PaidByUserID   uint    `json:"paidByUserId" gorm:"not null"`
	CategoryID     *uint   `json:"categoryId" gorm:"index"`
	ParticipantIDs string  `json:"-" gorm:"type:text;not null"` // JSON array of user IDs
	SplitRule      string  `json:"splitRule" gorm:"not null;default:equal"`
}

// ChargeResponse is returned to clients (participant IDs as a JSON array).
type ChargeResponse struct {
	ID             uint    `json:"id"`
	GroupID        uint    `json:"groupId"`
	Description    string  `json:"description"`
	Amount         float64 `json:"amount"`
	Date           string  `json:"date"`
	PaidByUserID   uint    `json:"paidByUserId"`
	CategoryID     *uint   `json:"categoryId"`
	CategoryName   string  `json:"categoryName,omitempty"`
	CategoryIcon   string  `json:"categoryIcon,omitempty"`
	ParticipantIDs []uint  `json:"participantIds"`
	SplitRule      string  `json:"splitRule"`
	SharePerPerson float64 `json:"sharePerPerson"`
}

// CreateGroupRequest is the body for POST /groups.
type CreateGroupRequest struct {
	Name     string `json:"name"`
	Currency string `json:"currency"`
}

// UpdateGroupRequest is the body for PUT /groups/:id.
type UpdateGroupRequest struct {
	Name     string `json:"name"`
	Currency string `json:"currency"`
}

// CreatePersonRequest is the body for POST /people.
type CreatePersonRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// UpdatePersonRequest is the body for PUT /people/:id.
type UpdatePersonRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// CreateCategoryRequest is the body for POST /categories.
type CreateCategoryRequest struct {
	Name string `json:"name"`
	Icon string `json:"icon"`
}

// UpdateCategoryRequest is the body for PUT /categories/:id.
type UpdateCategoryRequest struct {
	Name string `json:"name"`
	Icon string `json:"icon"`
}

// AddMemberRequest is the body for POST /groups/:id/members.
type AddMemberRequest struct {
	UserID uint `json:"userId"`
}

// CreateChargeRequest is the body for POST /groups/:id/charges.
type CreateChargeRequest struct {
	Description    string  `json:"description"`
	Amount         float64 `json:"amount"`
	Date           string  `json:"date"`
	PaidByUserID   uint    `json:"paidByUserId"`
	CategoryID     *uint   `json:"categoryId"`
	ParticipantIDs []uint  `json:"participantIds"`
	SplitRule      string  `json:"splitRule"`
}

// UpdateChargeRequest is the body for PUT /groups/:id/charges/:chargeId.
type UpdateChargeRequest struct {
	Description    string  `json:"description"`
	Amount         float64 `json:"amount"`
	Date           string  `json:"date"`
	PaidByUserID   uint    `json:"paidByUserId"`
	CategoryID     *uint   `json:"categoryId"`
	ParticipantIDs []uint  `json:"participantIds"`
	SplitRule      string  `json:"splitRule"`
}

// SettleRequest optionally limits which charges to include.
type SettleRequest struct {
	OnlyChargeIDs []uint `json:"onlyChargeIds"`
}

// BalanceRow is one member's net in a settlement.
type BalanceRow struct {
	UserID uint    `json:"userId"`
	Name   string  `json:"name"`
	Net    float64 `json:"net"`
}

// TransferRow is a simplified who-pays-whom payment.
type TransferRow struct {
	FromUserID uint    `json:"fromUserId"`
	FromName   string  `json:"fromName"`
	ToUserID   uint    `json:"toUserId"`
	ToName     string  `json:"toName"`
	Amount     float64 `json:"amount"`
}

// CategorySpendRow is total spend for one category within a settle.
type CategorySpendRow struct {
	CategoryID  *uint   `json:"categoryId"`
	Name        string  `json:"name"`
	Icon        string  `json:"icon"`
	Amount      float64 `json:"amount"`
	ChargeCount int     `json:"chargeCount"`
}

// SpendDayRow is total spend for one calendar day within a settle.
type SpendDayRow struct {
	Date        string  `json:"date"`
	Amount      float64 `json:"amount"`
	ChargeCount int     `json:"chargeCount"`
}

// SettleResponse is returned by POST /settle.
type SettleResponse struct {
	GroupID           uint               `json:"groupId"`
	Currency          string             `json:"currency"`
	Balances          []BalanceRow       `json:"balances"`
	Transfers         []TransferRow      `json:"transfers"`
	CategoryBreakdown []CategorySpendRow `json:"categoryBreakdown"`
	SpendOverTime     []SpendDayRow      `json:"spendOverTime"`
}

// ErrorBody is the standard API error shape.
type ErrorBody struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}
