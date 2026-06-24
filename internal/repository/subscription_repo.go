package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"subscription-service/internal/apperrors"
	"subscription-service/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SubscriptionRepository interface {
	Create(ctx context.Context, sub *model.Subscription) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error)
	List(ctx context.Context, userID *uuid.UUID, serviceName *string) ([]model.Subscription, error)
	ListActiveInPeriod(ctx context.Context, userID *uuid.UUID, serviceName *string, from, to time.Time) ([]model.Subscription, error)
	Update(ctx context.Context, sub *model.Subscription) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type PostgresSubscriptionRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresSubscriptionRepository(pool *pgxpool.Pool) *PostgresSubscriptionRepo {
	return &PostgresSubscriptionRepo{pool: pool}
}

const subscriptionColumns = `id, service_name, price, user_id, start_date, end_date`

func scanSubscription(row pgx.Row) (model.Subscription, error) {
	var sub model.Subscription
	err := row.Scan(
		&sub.ID,
		&sub.ServiceName,
		&sub.Price,
		&sub.UserID,
		&sub.StartDate,
		&sub.EndDate,
	)
	return sub, err
}

func (r *PostgresSubscriptionRepo) Create(ctx context.Context, sub *model.Subscription) error {
	query := `
		INSERT INTO subscriptions (id, service_name, price, user_id, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.pool.Exec(ctx, query,
		sub.ID, sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate,
	)
	if err != nil {
		return fmt.Errorf("create subscription: %w", err)
	}
	return nil
}

func (r *PostgresSubscriptionRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	query := `SELECT ` + subscriptionColumns + ` FROM subscriptions WHERE id = $1`

	sub, err := scanSubscription(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("get subscription: %w", err)
	}
	return &sub, nil
}

func (r *PostgresSubscriptionRepo) List(ctx context.Context, userID *uuid.UUID, serviceName *string) ([]model.Subscription, error) {
	query := `SELECT ` + subscriptionColumns + ` FROM subscriptions WHERE 1=1`
	args := []any{}
	argIdx := 1

	if userID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argIdx)
		args = append(args, *userID)
		argIdx++
	}
	if serviceName != nil {
		query += fmt.Sprintf(" AND service_name = $%d", argIdx)
		args = append(args, *serviceName)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list subscriptions: %w", err)
	}
	defer rows.Close()

	var subs []model.Subscription
	for rows.Next() {
		sub, err := scanSubscription(rows)
		if err != nil {
			return nil, fmt.Errorf("scan subscription: %w", err)
		}
		subs = append(subs, sub)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate subscriptions: %w", err)
	}

	return subs, nil
}

func (r *PostgresSubscriptionRepo) ListActiveInPeriod(
	ctx context.Context,
	userID *uuid.UUID,
	serviceName *string,
	from, to time.Time,
) ([]model.Subscription, error) {
	query := `SELECT ` + subscriptionColumns + ` FROM subscriptions WHERE start_date <= $2 AND (end_date IS NULL OR end_date >= $1)`
	args := []any{from, to}
	argIdx := 3

	if userID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argIdx)
		args = append(args, *userID)
		argIdx++
	}
	if serviceName != nil {
		query += fmt.Sprintf(" AND service_name = $%d", argIdx)
		args = append(args, *serviceName)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list active subscriptions: %w", err)
	}
	defer rows.Close()

	var subs []model.Subscription
	for rows.Next() {
		sub, err := scanSubscription(rows)
		if err != nil {
			return nil, fmt.Errorf("scan subscription: %w", err)
		}
		subs = append(subs, sub)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate subscriptions: %w", err)
	}

	return subs, nil
}

func (r *PostgresSubscriptionRepo) Update(ctx context.Context, sub *model.Subscription) error {
	query := `
		UPDATE subscriptions
		SET service_name = $2, price = $3, user_id = $4, start_date = $5, end_date = $6
		WHERE id = $1`

	tag, err := r.pool.Exec(ctx, query,
		sub.ID, sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate,
	)
	if err != nil {
		return fmt.Errorf("update subscription: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *PostgresSubscriptionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM subscriptions WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete subscription: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}
