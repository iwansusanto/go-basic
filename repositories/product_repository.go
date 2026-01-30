package repositories

import (
	"database/sql"
	"kasir-api/models"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// GetAll retrieves all active products
func (r *ProductRepository) GetAll() ([]models.Product, error) {
	rows, err := r.db.Query("SELECT id, name, price, stock, category_id, deleted_at FROM product WHERE deleted_at IS NULL")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		var deletedAt sql.NullTime
		if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Stock, &p.CategoryID, &deletedAt); err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			p.DeletedAt = timestamppb.New(deletedAt.Time)
		}
		products = append(products, p)
	}
	return products, nil
}

// GetByID retrieves a product by ID
func (r *ProductRepository) GetByID(id int) (models.Product, error) {
	var p models.Product
	var c models.Category
	var deletedAt sql.NullTime

	query := `
		SELECT p.id, p.name, p.price, p.stock, p.category_id, p.deleted_at, 
		       c.id, c.name, c.description
		FROM product p
		LEFT JOIN category c ON p.category_id = c.id
		WHERE p.id = $1 AND p.deleted_at IS NULL
	`

	err := r.db.QueryRow(query, id).Scan(
		&p.ID, &p.Name, &p.Price, &p.Stock, &p.CategoryID, &deletedAt,
		&c.ID, &c.Name, &c.Description,
	)

	if err != nil {
		return models.Product{}, err
	}

	p.Category = &c

	if deletedAt.Valid {
		p.DeletedAt = timestamppb.New(deletedAt.Time)
	}
	return p, nil
}

// Create inserts a new product
func (r *ProductRepository) Create(product models.Product) (models.Product, error) {
	var deletedAt sql.NullTime
	err := r.db.QueryRow(
		"INSERT INTO product (name, price, stock, category_id) VALUES ($1, $2, $3, $4) RETURNING id, deleted_at",
		product.Name, product.Price, product.Stock, product.CategoryID,
	).Scan(&product.ID, &deletedAt)

	if err != nil {
		return models.Product{}, err
	}

	if deletedAt.Valid {
		product.DeletedAt = timestamppb.New(deletedAt.Time)
	}
	return product, nil
}

// Update updates an existing product
func (r *ProductRepository) Update(product models.Product) (models.Product, error) {
	var deletedAt sql.NullTime
	err := r.db.QueryRow(
		"UPDATE product SET name = $1, price = $2, stock = $3, category_id = $4 WHERE id = $5 RETURNING deleted_at",
		product.Name, product.Price, product.Stock, product.CategoryID, product.ID,
	).Scan(&deletedAt)

	if err != nil {
		return models.Product{}, err
	}

	if deletedAt.Valid {
		product.DeletedAt = timestamppb.New(deletedAt.Time)
	}
	return product, nil
}

// Delete soft deletes a product
func (r *ProductRepository) Delete(id int) error {
	_, err := r.db.Exec("UPDATE product SET deleted_at = NOW() WHERE id = $1", id)
	return err
}
