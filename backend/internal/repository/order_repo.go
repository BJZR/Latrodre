package repository

import (
	"fmt"
	"strings"

	"latrode-fusion/internal/database"
	"latrode-fusion/internal/models"
)

type OrderRepo struct {
	db *database.DB
}

func NewOrderRepo(db *database.DB) *OrderRepo {
	return &OrderRepo{db: db}
}

func scanOrder(s scannable) (*models.Order, error) {
	o := &models.Order{}
	err := s.Scan(&o.ID, &o.UserID, &o.Total, &o.Status, &o.PaymentStatus,
		&o.PaymentMethod, &o.ShippingName, &o.ShippingPhone, &o.ShippingAddress,
		&o.ShippingCity, &o.ShippingPostalCode, &o.ShippingCountry, &o.CreatedAt)
	if err != nil {
		return nil, err
	}
	return o, nil
}

func (r *OrderRepo) FindByUserID(userID int) ([]models.Order, error) {
	rows, err := r.db.DB.Query(
		`SELECT id, user_id, total, status, payment_status, payment_method,
		 shipping_name, shipping_phone, shipping_address, shipping_city, shipping_postal_code, shipping_country, created_at
		 FROM orders WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		o.Items = r.findItems(o.ID)
		orders = append(orders, *o)
	}
	return orders, nil
}

func (r *OrderRepo) FindByID(id int) (*models.Order, error) {
	o, err := scanOrder(r.db.DB.QueryRow(
		`SELECT id, user_id, total, status, payment_status, payment_method,
		 shipping_name, shipping_phone, shipping_address, shipping_city, shipping_postal_code, shipping_country, created_at
		 FROM orders WHERE id=$1`, id))
	if err != nil {
		return nil, err
	}
	o.Items = r.findItems(o.ID)
	return o, nil
}

func (r *OrderRepo) Create(userID int, req *models.CreateOrderRequest,
	cartItems []models.CartItem, total float64) (*models.Order, error) {

	pm := req.PaymentMethod
	if pm == "" {
		pm = "cash_on_delivery"
	}
	o, err := scanOrder(r.db.DB.QueryRow(
		`INSERT INTO orders (user_id, total, status, payment_method,
		 shipping_name, shipping_phone, shipping_address, shipping_city, shipping_postal_code, shipping_country)
		 VALUES ($1, $2, 'pending', $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id, user_id, total, status, payment_status, payment_method,
		 shipping_name, shipping_phone, shipping_address, shipping_city, shipping_postal_code, shipping_country, created_at`,
		userID, total, pm, req.ShippingName, req.ShippingPhone, req.ShippingAddress,
		req.ShippingCity, req.ShippingPostalCode, req.ShippingCountry))
	if err != nil {
		return nil, err
	}

	for _, ci := range cartItems {
		subtotal := float64(ci.Quantity)
		if ci.Product != nil {
			subtotal *= ci.Product.Price
		}
		colorName := ""
		productName := ""
		productPrice := 0.0
		if ci.Product != nil {
			productName = ci.Product.Name
			productPrice = ci.Product.Price
		}
		if ci.Color != nil {
			colorName = ci.Color.Name
		}
		var productID *int
		if ci.Product != nil {
			productID = &ci.Product.ID
		}

		item := &models.OrderItem{}
		err := r.db.DB.QueryRow(
			`INSERT INTO order_items (order_id, product_id, product_name, product_price, color_name, size, quantity, subtotal)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			 RETURNING id, order_id, product_id, product_name, product_price, color_name, size, quantity, subtotal`,
			o.ID, productID, productName, productPrice, colorName, ci.Size, ci.Quantity, subtotal,
		).Scan(&item.ID, &item.OrderID, &item.ProductID, &item.ProductName, &item.ProductPrice,
			&item.ColorName, &item.Size, &item.Quantity, &item.Subtotal)
		if err != nil {
			return nil, err
		}
		o.Items = append(o.Items, *item)

		if ci.Product != nil {
			r.db.DB.Exec(
				`UPDATE products SET stock = stock - $1 WHERE id = $2 AND stock >= $1`,
				ci.Quantity, ci.Product.ID)
		}
		if ci.ColorID != nil && *ci.ColorID > 0 {
			r.db.DB.Exec(
				`UPDATE product_colors SET stock = stock - $1 WHERE id = $2 AND stock >= $1`,
				ci.Quantity, *ci.ColorID)
			if ci.Size != "" {
				r.db.DB.Exec(
					`UPDATE inventory SET stock = stock - $1 WHERE color_id = $2 AND size = $3 AND stock >= $1`,
					ci.Quantity, *ci.ColorID, ci.Size)
			}
		}
	}

	return o, nil
}

func (r *OrderRepo) UpdateStatus(id int, status, paymentStatus string) error {
	_, err := r.db.DB.Exec(
		`UPDATE orders SET status=$1, payment_status=$2, updated_at=NOW() WHERE id=$3`,
		status, paymentStatus, id)
	return err
}

func (r *OrderRepo) FindAllWithFilters(status, paymentStatus string) ([]models.Order, error) {
	query := `SELECT id, user_id, total, status, payment_status, payment_method,
	 shipping_name, shipping_phone, shipping_address, shipping_city, shipping_postal_code, shipping_country, created_at
	 FROM orders WHERE 1=1`
	var args []interface{}
	argIdx := 1

	if status != "" {
		query += fmt.Sprintf(" AND status=$%d", argIdx)
		args = append(args, status)
		argIdx++
	}
	if paymentStatus != "" {
		query += fmt.Sprintf(" AND payment_status=$%d", argIdx)
		args = append(args, paymentStatus)
		argIdx++
	}
	query += " ORDER BY created_at DESC"

	rows, err := r.db.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		o.Items = r.findItems(o.ID)
		orders = append(orders, *o)
	}
	return orders, nil
}

func (r *OrderRepo) FindRecent(limit int) ([]models.Order, error) {
	rows, err := r.db.DB.Query(
		`SELECT id, user_id, total, status, payment_status, payment_method,
		 shipping_name, shipping_phone, shipping_address, shipping_city, shipping_postal_code, shipping_country, created_at
		 FROM orders ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		o.Items = r.findItems(o.ID)
		orders = append(orders, *o)
	}
	return orders, nil
}

func (r *OrderRepo) GetDashboardStats() (*models.DashboardStats, error) {
	stats := &models.DashboardStats{}
	err := r.db.DB.QueryRow(
		`SELECT COALESCE(SUM(total), 0), COUNT(*),
		 COALESCE(SUM(CASE WHEN status='pending' THEN 1 ELSE 0 END), 0)
		 FROM orders`,
	).Scan(&stats.TotalRevenue, &stats.TotalOrders, &stats.PendingOrders)
	if err != nil {
		return nil, err
	}

	err = r.db.DB.QueryRow(
		`SELECT COUNT(*) FROM users WHERE role='customer'`,
	).Scan(&stats.TotalCustomers)
	if err != nil {
		return nil, err
	}

	stats.RecentOrders, _ = r.FindRecent(5)
	if stats.RecentOrders == nil {
		stats.RecentOrders = []models.Order{}
	}

	return stats, nil
}

func (r *OrderRepo) findItems(orderID int) []models.OrderItem {
	rows, err := r.db.DB.Query(
		`SELECT id, order_id, product_id, product_name, product_price, color_name, size, quantity, subtotal
		 FROM order_items WHERE order_id=$1`, orderID)
	if err != nil {
		return []models.OrderItem{}
	}
	defer rows.Close()

	var items []models.OrderItem
	for rows.Next() {
		var item models.OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.ProductName,
			&item.ProductPrice, &item.ColorName, &item.Size, &item.Quantity, &item.Subtotal); err != nil {
			continue
		}
		items = append(items, item)
	}
	if items == nil {
		items = []models.OrderItem{}
	}
	return items
}

func (r *OrderRepo) LogActivity(userID int, action, entity string, entityID int, ip string) error {
	var uid *int
	if userID > 0 {
		uid = &userID
	}
	_, err := r.db.DB.Exec(
		`INSERT INTO activity_logs (user_id, action, entity, entity_id, ip_address) VALUES ($1, $2, $3, $4, $5)`,
		uid, action, entity, entityID, ip)
	return err
}

func (r *OrderRepo) GetLogs() ([]models.ActivityLog, error) {
	rows, err := r.db.DB.Query(
		`SELECT id, user_id, action, entity, entity_id, ip_address, created_at
		 FROM activity_logs ORDER BY created_at DESC LIMIT 100`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.ActivityLog
	for rows.Next() {
		var l models.ActivityLog
		if err := rows.Scan(&l.ID, &l.UserID, &l.Action, &l.Entity, &l.EntityID, &l.IPAddress, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}

func (r *OrderRepo) GetPaymentMethods() ([]models.PaymentMethod, error) {
	rows, err := r.db.DB.Query(
		`SELECT id, name, description, enabled, created_at FROM payment_methods ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var methods []models.PaymentMethod
	for rows.Next() {
		var m models.PaymentMethod
		if err := rows.Scan(&m.ID, &m.Name, &m.Description, &m.Enabled, &m.CreatedAt); err != nil {
			return nil, err
		}
		methods = append(methods, m)
	}
	return methods, nil
}

func (r *OrderRepo) GetSettings() ([]models.Setting, error) {
	rows, err := r.db.DB.Query(
		`SELECT id, key, value FROM settings ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []models.Setting
	for rows.Next() {
		var s models.Setting
		if err := rows.Scan(&s.ID, &s.Key, &s.Value); err != nil {
			return nil, err
		}
		settings = append(settings, s)
	}
	return settings, nil
}

func (r *OrderRepo) UpdateSetting(key, value string) error {
	_, err := r.db.DB.Exec(
		`INSERT INTO settings (key, value) VALUES ($1, $2)
		 ON CONFLICT (key) DO UPDATE SET value=$2, updated_at=NOW()`, key, value)
	return err
}

func buildQuery(base string, conditions []string, args []interface{}) (string, []interface{}) {
	if len(conditions) > 0 {
		base += " WHERE " + strings.Join(conditions, " AND ")
	}
	return base, args
}
