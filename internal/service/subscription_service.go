package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"subscription-service/internal/apperrors"
	"subscription-service/internal/model"
	"subscription-service/internal/repository"

	"github.com/google/uuid"
)

type SubscriptionService interface {
	CreateSubscription(ctx context.Context, req model.CreateSubscriptionRequest) (*model.SubscriptionResponse, error)
	GetSubscription(ctx context.Context, id uuid.UUID) (*model.SubscriptionResponse, error)
	ListSubscriptions(ctx context.Context, userID *uuid.UUID, serviceName *string) ([]model.SubscriptionResponse, error)
	UpdateSubscription(ctx context.Context, id uuid.UUID, req model.UpdateSubscriptionRequest) (*model.SubscriptionResponse, error)
	DeleteSubscription(ctx context.Context, id uuid.UUID) error
	CalculateTotalSum(ctx context.Context, req model.SumRequest) (int, error)
}

type subscriptionService struct {
	repo   repository.SubscriptionRepository
	logger *slog.Logger
}

func NewSubscriptionService(repo repository.SubscriptionRepository, logger *slog.Logger) SubscriptionService {
	return &subscriptionService{repo: repo, logger: logger}
}

func (s *subscriptionService) CreateSubscription(ctx context.Context, req model.CreateSubscriptionRequest) (*model.SubscriptionResponse, error) {
	sub, err := requestToSubscription(uuid.New(), req.ServiceName, req.Price, req.UserID, req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, sub); err != nil {
		s.logger.Error("create subscription", "error", err)
		return nil, fmt.Errorf("create subscription: %w", err)
	}

	return toResponse(sub), nil
}

func (s *subscriptionService) GetSubscription(ctx context.Context, id uuid.UUID) (*model.SubscriptionResponse, error) {
	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, apperrors.ErrNotFound
		}
		s.logger.Error("get subscription", "id", id, "error", err)
		return nil, fmt.Errorf("get subscription: %w", err)
	}

	return toResponse(sub), nil
}

func (s *subscriptionService) ListSubscriptions(ctx context.Context, userID *uuid.UUID, serviceName *string) ([]model.SubscriptionResponse, error) {
	subs, err := s.repo.List(ctx, userID, serviceName)
	if err != nil {
		s.logger.Error("list subscriptions", "error", err)
		return nil, fmt.Errorf("list subscriptions: %w", err)
	}

	resp := make([]model.SubscriptionResponse, 0, len(subs))
	for i := range subs {
		resp = append(resp, *toResponse(&subs[i]))
	}
	return resp, nil
}

func (s *subscriptionService) UpdateSubscription(ctx context.Context, id uuid.UUID, req model.UpdateSubscriptionRequest) (*model.SubscriptionResponse, error) {
	sub, err := requestToSubscription(id, req.ServiceName, req.Price, req.UserID, req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, sub); err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, apperrors.ErrNotFound
		}
		s.logger.Error("update subscription", "id", id, "error", err)
		return nil, fmt.Errorf("update subscription: %w", err)
	}

	return toResponse(sub), nil
}

func (s *subscriptionService) DeleteSubscription(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return apperrors.ErrNotFound
		}
		s.logger.Error("delete subscription", "id", id, "error", err)
		return fmt.Errorf("delete subscription: %w", err)
	}
	return nil
}

func (s *subscriptionService) CalculateTotalSum(ctx context.Context, req model.SumRequest) (int, error) {
	from, err := parseMonthYear(req.From)
	if err != nil {
		return 0, err
	}

	to, err := parseMonthYear(req.To)
	if err != nil {
		return 0, err
	}

	if from.After(to) {
		return 0, fmt.Errorf("%w: from must be before or equal to to", apperrors.ErrInvalidDate)
	}

	subs, err := s.repo.ListActiveInPeriod(ctx, req.UserID, req.ServiceName, from, to)
	if err != nil {
		s.logger.Error("calculate total sum", "error", err)
		return 0, fmt.Errorf("calculate total sum: %w", err)
	}

	total := 0
	for i := range subs {
		months := model.OverlappingMonths(subs[i].StartDate, subs[i].EndDate, from, to)
		total += subs[i].Price * months
	}

	return total, nil
}

func requestToSubscription(
	id uuid.UUID,
	serviceName string,
	price int,
	userID uuid.UUID,
	startDate string,
	endDate *string,
) (*model.Subscription, error) {
	start, err := parseMonthYear(startDate)
	if err != nil {
		return nil, err
	}

	var end *time.Time
	if endDate != nil {
		parsed, err := parseMonthYear(*endDate)
		if err != nil {
			return nil, err
		}
		if parsed.Before(start) {
			return nil, fmt.Errorf("%w: end_date must be after or equal to start_date", apperrors.ErrInvalidDate)
		}
		end = &parsed
	}

	return &model.Subscription{
		ID:          id,
		ServiceName: serviceName,
		Price:       price,
		UserID:      userID,
		StartDate:   start,
		EndDate:     end,
	}, nil
}

func parseMonthYear(value string) (time.Time, error) {
	t, err := model.ParseMonthYear(value)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: %v", apperrors.ErrInvalidDate, err)
	}
	return t, nil
}

func toResponse(sub *model.Subscription) *model.SubscriptionResponse {
	resp := &model.SubscriptionResponse{
		ID:          sub.ID,
		ServiceName: sub.ServiceName,
		Price:       sub.Price,
		UserID:      sub.UserID,
		StartDate:   model.FormatMonthYear(sub.StartDate),
	}

	if sub.EndDate != nil {
		end := model.FormatMonthYear(*sub.EndDate)
		resp.EndDate = &end
	}

	return resp
}
