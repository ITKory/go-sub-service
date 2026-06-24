package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"subscription-service/internal/apperrors"
	"subscription-service/internal/model"
	"subscription-service/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type SubscriptionHandler struct {
	service  service.SubscriptionService
	logger   *slog.Logger
	validate *validator.Validate
}

func NewSubscriptionHandler(svc service.SubscriptionService, logger *slog.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{
		service:  svc,
		logger:   logger,
		validate: validator.New(),
	}
}

// CreateSubscription godoc
//
//	@Summary		Create
//	@Tags			subscriptions
//	@Accept			json
//	@Produce		json
//	@Param			input	body		model.CreateSubscriptionRequest	true	"Данные подписки"
//	@Success		201		{object}	model.SubscriptionResponse
//	@Failure		400		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/subscriptions [post]
func (h *SubscriptionHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var req model.CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("invalid json", "error", err)
		h.sendError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.service.CreateSubscription(r.Context(), req)
	if err != nil {
		h.handleServiceError(w, err, "create subscription")
		return
	}

	h.sendJSON(w, http.StatusCreated, resp)
}

// GetSubscription godoc
//
//	@Summary		Get by id
//	@Tags			subscriptions
//	@Produce		json
//	@Param			id	path		string	true	"UUID подписки"
//	@Success		200	{object}	model.SubscriptionResponse
//	@Failure		404	{object}	map[string]string
//	@Router			/subscriptions/{id} [get]
func (h *SubscriptionHandler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r.PathValue("id"))
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid id")
		return
	}

	resp, err := h.service.GetSubscription(r.Context(), id)
	if err != nil {
		h.handleServiceError(w, err, "get subscription")
		return
	}

	h.sendJSON(w, http.StatusOK, resp)
}

// ListSubscriptions godoc
//
//	@Summary		Subscription list
//	@Tags			subscriptions
//	@Produce		json
//	@Param			user_id			query	string	false	"ID пользователя"
//	@Param			service_name	query	string	false	"Название сервиса"
//	@Success		200	{array}	model.SubscriptionResponse
//	@Router			/subscriptions [get]
func (h *SubscriptionHandler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	var userID *uuid.UUID
	if uidStr := r.URL.Query().Get("user_id"); uidStr != "" {
		uid, err := parseUUID(uidStr)
		if err != nil {
			h.sendError(w, http.StatusBadRequest, "invalid user_id")
			return
		}
		userID = &uid
	}

	var serviceName *string
	if name := r.URL.Query().Get("service_name"); name != "" {
		serviceName = &name
	}

	resp, err := h.service.ListSubscriptions(r.Context(), userID, serviceName)
	if err != nil {
		h.handleServiceError(w, err, "list subscriptions")
		return
	}

	h.sendJSON(w, http.StatusOK, resp)
}

// UpdateSubscription godoc
//
//	@Summary		Обновить подписку
//	@Tags			subscriptions
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"UUID подписки"
//	@Param			input	body		model.UpdateSubscriptionRequest	true	"Данные подписки"
//	@Success		200		{object}	model.SubscriptionResponse
//	@Failure		400		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Router			/subscriptions/{id} [put]
func (h *SubscriptionHandler) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r.PathValue("id"))
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req model.UpdateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.service.UpdateSubscription(r.Context(), id, req)
	if err != nil {
		h.handleServiceError(w, err, "update subscription")
		return
	}

	h.sendJSON(w, http.StatusOK, resp)
}

// DeleteSubscription godoc
//
//	@Summary		Удалить подписку
//	@Tags			subscriptions
//	@Param			id	path	string	true	"UUID подписки"
//	@Success		204
//	@Failure		404	{object}	map[string]string
//	@Router			/subscriptions/{id} [delete]
func (h *SubscriptionHandler) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r.PathValue("id"))
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.service.DeleteSubscription(r.Context(), id); err != nil {
		h.handleServiceError(w, err, "delete subscription")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CalculateTotalSum godoc
//
//	@Summary		Сумма подписок за период
//	@Description	Считает стоимость по числу месяцев, когда подписка была активна в указанном диапазоне
//	@Tags			subscriptions
//	@Accept			json
//	@Produce		json
//	@Param			input	body		model.SumRequest	true	"Параметры"
//	@Success		200		{object}	model.SumResponse
//	@Failure		400		{object}	map[string]string
//	@Router			/subscriptions/sum [post]
func (h *SubscriptionHandler) CalculateTotalSum(w http.ResponseWriter, r *http.Request) {
	var req model.SumRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		h.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	total, err := h.service.CalculateTotalSum(r.Context(), req)
	if err != nil {
		h.handleServiceError(w, err, "calculate total sum")
		return
	}

	h.sendJSON(w, http.StatusOK, model.SumResponse{TotalSum: total})
}

func (h *SubscriptionHandler) handleServiceError(w http.ResponseWriter, err error, action string) {
	switch {
	case errors.Is(err, apperrors.ErrNotFound):
		h.sendError(w, http.StatusNotFound, "subscription not found")
	case errors.Is(err, apperrors.ErrInvalidDate):
		h.sendError(w, http.StatusBadRequest, err.Error())
	default:
		h.logger.Error(action, "error", err)
		h.sendError(w, http.StatusInternalServerError, "internal server error")
	}
}

func parseUUID(value string) (uuid.UUID, error) {
	return uuid.Parse(value)
}

func (h *SubscriptionHandler) sendJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("encode response", "error", err)
	}
}

func (h *SubscriptionHandler) sendError(w http.ResponseWriter, status int, message string) {
	h.sendJSON(w, status, map[string]string{"error": message})
}
