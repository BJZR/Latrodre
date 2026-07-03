package handlers

import (
	"net/http"

	"latrode-fusion/internal/middleware"
	"latrode-fusion/internal/models"
	"latrode-fusion/internal/repository"
)

type PaymentHandler struct {
	orderRepo *repository.OrderRepo
}

func NewPaymentHandler(orderRepo *repository.OrderRepo) *PaymentHandler {
	return &PaymentHandler{orderRepo: orderRepo}
}

func (h *PaymentHandler) ListMethods(w http.ResponseWriter, r *http.Request) {
	methods, err := h.orderRepo.GetPaymentMethods()
	if err != nil {
		middleware.WriteJSON(w, http.StatusInternalServerError, middleware.APIError{Error: "error al obtener métodos de pago"})
		return
	}

	if methods == nil {
		methods = []models.PaymentMethod{}
	}

	middleware.WriteJSON(w, http.StatusOK, methods)
}
