package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"latrode-fusion/internal/middleware"
	"latrode-fusion/internal/models"
	"latrode-fusion/internal/repository"
)

type AdminHandler struct {
	orderRepo   *repository.OrderRepo
	productRepo *repository.ProductRepo
	userRepo    *repository.UserRepo
}

func NewAdminHandler(orderRepo *repository.OrderRepo, productRepo *repository.ProductRepo, userRepo *repository.UserRepo) *AdminHandler {
	return &AdminHandler{orderRepo: orderRepo, productRepo: productRepo, userRepo: userRepo}
}

func (h *AdminHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	stats, err := h.orderRepo.GetDashboardStats()
	if err != nil {
		middleware.WriteJSON(w, http.StatusInternalServerError, middleware.APIError{Error: "error al obtener estadísticas"})
		return
	}

	topProducts, err := h.productRepo.FindTopSelling(5)
	if err == nil && topProducts != nil {
		stats.TopProducts = topProducts
	}
	if stats.TopProducts == nil {
		stats.TopProducts = []models.Product{}
	}

	middleware.WriteJSON(w, http.StatusOK, stats)
}

func (h *AdminHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	paymentStatus := r.URL.Query().Get("payment_status")

	orders, err := h.orderRepo.FindAllWithFilters(status, paymentStatus)
	if err != nil {
		middleware.WriteJSON(w, http.StatusInternalServerError, middleware.APIError{Error: "error al obtener órdenes"})
		return
	}

	if orders == nil {
		orders = []models.Order{}
	}

	middleware.WriteJSON(w, http.StatusOK, orders)
}

func (h *AdminHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		middleware.WriteJSON(w, http.StatusBadRequest, middleware.APIError{Error: "id inválido"})
		return
	}

	var req models.UpdateOrderStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSON(w, http.StatusBadRequest, middleware.APIError{Error: "datos inválidos"})
		return
	}

	if err := h.orderRepo.UpdateStatus(id, req.Status, req.PaymentStatus); err != nil {
		middleware.WriteJSON(w, http.StatusInternalServerError, middleware.APIError{Error: "error al actualizar orden"})
		return
	}

	user := middleware.GetUser(r)
	h.orderRepo.LogActivity(user.ID, "actualizar_orden", "orders", id, middleware.GetClientIP(r))

	order, _ := h.orderRepo.FindByID(id)
	middleware.WriteJSON(w, http.StatusOK, order)
}

func (h *AdminHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.productRepo.FindAll()
	if err != nil {
		middleware.WriteJSON(w, http.StatusInternalServerError, middleware.APIError{Error: "error al obtener productos"})
		return
	}

	if products == nil {
		products = []models.Product{}
	}

	middleware.WriteJSON(w, http.StatusOK, products)
}

func (h *AdminHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req models.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSON(w, http.StatusBadRequest, middleware.APIError{Error: "datos inválidos"})
		return
	}

	product, err := h.productRepo.Create(&req)
	if err != nil {
		middleware.WriteJSON(w, http.StatusInternalServerError, middleware.APIError{Error: "error al crear producto"})
		return
	}

	user := middleware.GetUser(r)
	h.orderRepo.LogActivity(user.ID, "crear_producto", "products", product.ID, middleware.GetClientIP(r))

	middleware.WriteJSON(w, http.StatusCreated, product)
}

func (h *AdminHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		middleware.WriteJSON(w, http.StatusBadRequest, middleware.APIError{Error: "id inválido"})
		return
	}

	var req models.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSON(w, http.StatusBadRequest, middleware.APIError{Error: "datos inválidos"})
		return
	}

	product, err := h.productRepo.Update(id, &req)
	if err != nil {
		middleware.WriteJSON(w, http.StatusInternalServerError, middleware.APIError{Error: "error al actualizar producto"})
		return
	}

	user := middleware.GetUser(r)
	h.orderRepo.LogActivity(user.ID, "actualizar_producto", "products", id, middleware.GetClientIP(r))

	middleware.WriteJSON(w, http.StatusOK, product)
}

func (h *AdminHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		middleware.WriteJSON(w, http.StatusBadRequest, middleware.APIError{Error: "id inválido"})
		return
	}

	if err := h.productRepo.Delete(id); err != nil {
		middleware.WriteJSON(w, http.StatusInternalServerError, middleware.APIError{Error: "error al eliminar producto"})
		return
	}

	user := middleware.GetUser(r)
	h.orderRepo.LogActivity(user.ID, "eliminar_producto", "products", id, middleware.GetClientIP(r))

	middleware.WriteJSON(w, http.StatusOK, map[string]string{"message": "producto eliminado"})
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userRepo.FindAll()
	if err != nil {
		middleware.WriteJSON(w, http.StatusInternalServerError, middleware.APIError{Error: "error al obtener usuarios"})
		return
	}

	if users == nil {
		users = []models.User{}
	}

	middleware.WriteJSON(w, http.StatusOK, users)
}

func (h *AdminHandler) GetPaymentMethods(w http.ResponseWriter, r *http.Request) {
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

func (h *AdminHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.orderRepo.GetSettings()
	if err != nil {
		middleware.WriteJSON(w, http.StatusInternalServerError, middleware.APIError{Error: "error al obtener configuraciones"})
		return
	}

	if settings == nil {
		settings = []models.Setting{}
	}

	middleware.WriteJSON(w, http.StatusOK, settings)
}

func (h *AdminHandler) UpdateSetting(w http.ResponseWriter, r *http.Request) {
	var req models.Setting
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.WriteJSON(w, http.StatusBadRequest, middleware.APIError{Error: "datos inválidos"})
		return
	}

	if err := h.orderRepo.UpdateSetting(req.Key, req.Value); err != nil {
		middleware.WriteJSON(w, http.StatusInternalServerError, middleware.APIError{Error: "error al actualizar configuración"})
		return
	}

	user := middleware.GetUser(r)
	h.orderRepo.LogActivity(user.ID, "actualizar_config", "settings", 0, middleware.GetClientIP(r))

	middleware.WriteJSON(w, http.StatusOK, map[string]string{"message": "configuración actualizada"})
}

func (h *AdminHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
	logs, err := h.orderRepo.GetLogs()
	if err != nil {
		middleware.WriteJSON(w, http.StatusInternalServerError, middleware.APIError{Error: "error al obtener logs"})
		return
	}

	if logs == nil {
		logs = []models.ActivityLog{}
	}

	middleware.WriteJSON(w, http.StatusOK, logs)
}
