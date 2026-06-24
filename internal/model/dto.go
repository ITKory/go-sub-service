package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type CreateSubscriptionRequest struct {
	ServiceName string    `json:"service_name" validate:"required"`
	Price       int       `json:"price" validate:"required,gte=0"`
	UserID      uuid.UUID `json:"user_id" validate:"required"`
	StartDate   string    `json:"start_date" validate:"required"`
	EndDate     *string   `json:"end_date,omitempty"`
}

type UpdateSubscriptionRequest struct {
	ServiceName string    `json:"service_name" validate:"required"`
	Price       int       `json:"price" validate:"required,gte=0"`
	UserID      uuid.UUID `json:"user_id" validate:"required"`
	StartDate   string    `json:"start_date" validate:"required"`
	EndDate     *string   `json:"end_date,omitempty"`
}

type SubscriptionResponse struct {
	ID          uuid.UUID `json:"id"`
	ServiceName string    `json:"service_name"`
	Price       int       `json:"price"`
	UserID      uuid.UUID `json:"user_id"`
	StartDate   string    `json:"start_date"`
	EndDate     *string   `json:"end_date,omitempty"`
}

type SumRequest struct {
	UserID      *uuid.UUID `json:"user_id,omitempty"`
	ServiceName *string    `json:"service_name,omitempty"`
	From        string     `json:"from" validate:"required"`
	To          string     `json:"to" validate:"required"`
}

type SumResponse struct {
	TotalSum int `json:"total_sum"`
}

func ParseMonthYear(dateStr string) (time.Time, error) {
	t, err := time.Parse("01-2006", dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("expected MM-YYYY: %w", err)
	}
	return t, nil
}

func FormatMonthYear(t time.Time) string {
	return t.Format("01-2006")
}

// OverlappingMonths returns how many calendar months of [from, to] overlap with
// the subscription period. Price is monthly, so total cost = price * months.
func OverlappingMonths(subStart time.Time, subEnd *time.Time, from, to time.Time) int {
	periodEnd := to
	if subEnd != nil && subEnd.Before(periodEnd) {
		periodEnd = *subEnd
	}

	periodStart := subStart
	if from.After(periodStart) {
		periodStart = from
	}

	if periodStart.After(periodEnd) {
		return 0
	}

	return (periodEnd.Year()-periodStart.Year())*12 + int(periodEnd.Month()-periodStart.Month()) + 1
}
