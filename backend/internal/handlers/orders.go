package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"latrode-fusion/internal/middleware"
	"latrode-fusion/internal/models"
	"latrode-fusion/internal/repository"
)

type OrderHandler struct {
	orderRepo *repository.OrderRepo
	cartRepo  *repository.CartRepo
}

func NewOrderHandler(orderRepo *repository.OrderRepo, cartRepo *repository.CartRepo) *OrderHandler {
	return &OrderHandler{orderRepo: orderRepo, cartRepo: cartRepo}
}

func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	if user == nil {
		middleware.WriteJSON(w, http.StatusUnauthorized, middleware.APIError{Error: "debes iniciar sesión"})
		return
	}

	var req models.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSON(w, http.StatusBadRequest, middleware.APIError{Error: "datos inválidos"})
		return
	}

	cartItems, err := h.cartRepo.FindByUserID(user.ID)
	if err != nil {
		middleware.WriteJSON(w, http.StatusInternalServerError, middleware.APIError{Error: "error al obtener carrito"})
		return
	}

	if len(cartItems) == 0 {
		middleware.WriteJSON(w, http.StatusBadRequest, middleware.APIError{Error: "el carrito está vacío"})
		return
	}

	var total float64
	for _, item := range cartItems {
		if item.Product != nil {
			total += item.Product.Price * float64(item.Quantity)
		}
	}

	order, err := h.orderRepo.Create(user.ID, &req, cartItems, total)
	if err != nil {
		middleware.WriteJSON(w, http.StatusInternalServerError, middleware.APIError{Error: "error al crear orden"})
		return
	}

	h.cartRepo.Clear(user.ID)
	h.orderRepo.LogActivity(user.ID, "crear_orden", "orders", order.ID, middleware.GetClientIP(r))

	middleware.WriteJSON(w, http.StatusCreated, order)
}

func (h *OrderHandler) GetMyOrders(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	if user == nil {
		middleware.WriteJSON(w, http.StatusUnauthorized, middleware.APIError{Error: "no autorizado"})
		return
	}

	orders, err := h.orderRepo.FindByUserID(user.ID)
	if err != nil {
		middleware.WriteJSON(w, http.StatusInternalServerError, middleware.APIError{Error: "error al obtener órdenes"})
		return
	}

	if orders == nil {
		orders = []models.Order{}
	}

	middleware.WriteJSON(w, http.StatusOK, orders)
}

func (h *OrderHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		middleware.WriteJSON(w, http.StatusBadRequest, middleware.APIError{Error: "id inválido"})
		return
	}

	order, err := h.orderRepo.FindByID(id)
	if err != nil {
		middleware.WriteJSON(w, http.StatusNotFound, middleware.APIError{Error: "orden no encontrada"})
		return
	}

	middleware.WriteJSON(w, http.StatusOK, order)
}
