package models

import "time"

// Group is one split session (trip, apartment month, etc.).
type Group struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"not null"`
	Currency  string    `json:"currency" gorm:"not null;default:USD"`
	CreatedAt time.Time `json:"createdAt"`
}

// Person belongs to a group.
type Person struct {
	ID      uint   `json:"id" gorm:"primaryKey"`
	GroupID uint   `json:"groupId" gorm:"index;not null"`
	Name    string `json:"name" gorm:"not null"`
	Email   string `json:"email"`
}

// Charge is a shared expense within a group.
// ParticipantIDs is stored as JSON text (who shares this charge).
type Charge struct {
	ID              uint    `json:"id" gorm:"primaryKey"`
	GroupID         uint    `json:"groupId" gorm:"index;not null"`
	Description     string  `json:"description" gorm:"not null"`
	Amount          float64 `json:"amount" gorm:"not null"`
	PaidByPersonID  uint    `json:"paidByPersonId" gorm:"not null"`
	ParticipantIDs  string  `json:"-" gorm:"type:text;not null"` // JSON array of IDs
	SplitRule       string  `json:"splitRule" gorm:"not null;default:equal"`
}

// ChargeResponse is returned to clients (participant IDs as a JSON array).
type ChargeResponse struct {
	ID             uint    `json:"id"`
	GroupID        uint    `json:"groupId"`
	Description    string  `json:"description"`
	Amount         float64 `json:"amount"`
	PaidByPersonID uint    `json:"paidByPersonId"`
	ParticipantIDs []uint  `json:"participantIds"`
	SplitRule      string  `json:"splitRule"`
	SharePerPerson float64 `json:"sharePerPerson"`
}

// CreateGroupRequest is the body for POST /groups.
type CreateGroupRequest struct {
	Name     string `json:"name"`
	Currency string `json:"currency"`
}

// CreatePersonRequest is the body for POST /groups/:id/people.
type CreatePersonRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// CreateChargeRequest is the body for POST /groups/:id/charges.
type CreateChargeRequest struct {
	Description    string  `json:"description"`
	Amount         float64 `json:"amount"`
	PaidByPersonID uint    `json:"paidByPersonId"`
	ParticipantIDs []uint  `json:"participantIds"`
	SplitRule      string  `json:"splitRule"`
}

// SettleRequest optionally limits which charges to include.
type SettleRequest struct {
	OnlyChargeIDs []uint `json:"onlyChargeIds"`
}

// BalanceRow is one person's net in a settlement.
type BalanceRow struct {
	PersonID uint    `json:"personId"`
	Name     string  `json:"name"`
	Net      float64 `json:"net"`
}

// TransferRow is a simplified who-pays-whom payment.
type TransferRow struct {
	FromPersonID uint    `json:"fromPersonId"`
	FromName     string  `json:"fromName"`
	ToPersonID   uint    `json:"toPersonId"`
	ToName       string  `json:"toName"`
	Amount       float64 `json:"amount"`
}

// SettleResponse is returned by POST /settle.
type SettleResponse struct {
	GroupID   uint          `json:"groupId"`
	Currency  string        `json:"currency"`
	Balances  []BalanceRow  `json:"balances"`
	Transfers []TransferRow `json:"transfers"`
}

// ErrorBody is the standard API error shape.
type ErrorBody struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}
