package repository

import (
	"fmt"

	"latrode-fusion/internal/database"
	"latrode-fusion/internal/models"
)

type CartRepo struct {
	db *database.DB
}

func NewCartRepo(db *database.DB) *CartRepo {
	return &CartRepo{db: db}
}

func (r *CartRepo) FindByUserID(userID int) ([]models.CartItem, error) {
	rows, err := r.db.DB.Query(
		`SELECT ci.id, ci.user_id, ci.product_id, ci.color_id, ci.size, ci.quantity, ci.created_at,
		        COALESCE(pc.id, 0), COALESCE(pc.name, ''), COALESCE(pc.hex, '')
		 FROM cart_items ci
		 LEFT JOIN product_colors pc ON pc.id = ci.color_id
		 WHERE ci.user_id=$1 ORDER BY ci.created_at`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.CartItem
	for rows.Next() {
		var item models.CartItem
		var colorID int
		var colorName, colorHex string
		err := rows.Scan(&item.ID, &item.UserID, &item.ProductID, &item.ColorID,
			&item.Size, &item.Quantity, &item.CreatedAt,
			&colorID, &colorName, &colorHex)
		if err != nil {
			return nil, err
		}
		if colorName != "" && item.ColorID != nil {
			item.Color = &models.Color{
				ID:   colorID,
				Name: colorName,
				Hex:  colorHex,
			}
		}
		items = append(items, item)
	}

	for i := range items {
		items[i].Product = r.getProduct(items[i].ProductID)
	}

	return items, nil
}

func (r *CartRepo) AddItem(userID, productID, colorID int, size string, quantity int) error {
	var err error
	if colorID > 0 {
		var colorStock int
		err = r.db.DB.QueryRow(`SELECT stock FROM product_colors WHERE id=$1`, colorID).Scan(&colorStock)
		if err != nil {
			return fmt.Errorf("color no encontrado")
		}

		var existingQty int
		r.db.DB.QueryRow(
			`SELECT COALESCE(SUM(quantity), 0) FROM cart_items WHERE user_id=$1 AND product_id=$2 AND color_id=$3`,
			userID, productID, colorID).Scan(&existingQty)

		if existingQty+quantity > colorStock {
			available := colorStock - existingQty
			if available < 0 {
				available = 0
			}
			return fmt.Errorf("stock insuficiente: solo %d disponibles para este color", available)
		}
	} else {
		var stock int
		err := r.db.DB.QueryRow(`SELECT stock FROM products WHERE id=$1`, productID).Scan(&stock)
		if err != nil {
			return fmt.Errorf("producto no encontrado")
		}
		var existingQty int
		r.db.DB.QueryRow(
			`SELECT COALESCE(SUM(quantity), 0) FROM cart_items WHERE user_id=$1 AND product_id=$2`,
			userID, productID).Scan(&existingQty)
		if existingQty+quantity > stock {
			available := stock - existingQty
			if available < 0 {
				available = 0
			}
			return fmt.Errorf("stock insuficiente: solo %d disponibles", available)
		}
	}

	var existingID int
	err = r.db.DB.QueryRow(
		`SELECT id FROM cart_items WHERE user_id=$1 AND product_id=$2 AND color_id=$3 AND size=$4`,
		userID, productID, colorID, size,
	).Scan(&existingID)

	if err == nil {
		_, err = r.db.DB.Exec(
			`UPDATE cart_items SET quantity = quantity + $1 WHERE id=$2`, quantity, existingID)
		return err
	}

	_, err = r.db.DB.Exec(
		`INSERT INTO cart_items (user_id, product_id, color_id, size, quantity)
		 VALUES ($1, $2, $3, $4, $5)`,
		userID, productID, colorID, size, quantity)
	return err
}

func (r *CartRepo) UpdateQuantity(itemID, userID, quantity int) error {
	_, err := r.db.DB.Exec(
		`UPDATE cart_items SET quantity=$1 WHERE id=$2 AND user_id=$3`,
		quantity, itemID, userID)
	return err
}

func (r *CartRepo) RemoveItem(itemID, userID int) error {
	_, err := r.db.DB.Exec(
		`DELETE FROM cart_items WHERE id=$1 AND user_id=$2`, itemID, userID)
	return err
}

func (r *CartRepo) Clear(userID int) error {
	_, err := r.db.DB.Exec(`DELETE FROM cart_items WHERE user_id=$1`, userID)
	return err
}

func (r *CartRepo) getProduct(productID int) *models.Product {
	p, err := scanProduct(r.db.DB.QueryRow(
		`SELECT id, name, description, price, stock, category, image_url, sizes, material, care, created_at
		 FROM products WHERE id=$1`, productID))
	if err != nil {
		return nil
	}
	return p
}
