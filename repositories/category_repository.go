package repositories

import (
	"database/sql"
	"kasir-api/models"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type CategoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

// GetCategories retrieves all active categories from the database
func (r *CategoryRepository) GetAll() ([]models.Category, error) {
	rows, err := r.db.Query("SELECT id, name, description, deleted_at FROM category WHERE deleted_at IS NULL")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var c models.Category
		var deletedAt sql.NullTime
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &deletedAt); err != nil {
			return nil, err
		}
		if deletedAt.Valid {
			c.DeletedAt = timestamppb.New(deletedAt.Time)
		}
		categories = append(categories, c)
	}

	return categories, nil
}

// Create inserts a new category into the database
func (r *CategoryRepository) Create(category models.Category) (models.Category, error) {
	var deletedAt sql.NullTime
	err := r.db.QueryRow(
		"INSERT INTO category (name, description) VALUES ($1, $2) RETURNING id, deleted_at",
		category.Name, category.Description,
	).Scan(&category.ID, &deletedAt)

	if err != nil {
		return models.Category{}, err
	}

	if deletedAt.Valid {
		category.DeletedAt = timestamppb.New(deletedAt.Time)
	}

	return category, nil
}

// GetByID retrieves a category by its ID
func (r *CategoryRepository) GetByID(id int) (models.Category, error) {
	var c models.Category
	var deletedAt sql.NullTime
	err := r.db.QueryRow(
		"SELECT id, name, description, deleted_at FROM category WHERE id = $1 AND deleted_at IS NULL",
		id,
	).Scan(&c.ID, &c.Name, &c.Description, &deletedAt)

	if err != nil {
		return models.Category{}, err
	}

	if deletedAt.Valid {
		c.DeletedAt = timestamppb.New(deletedAt.Time)
	}

	return c, nil
}

// Delete soft deletes a category by its ID
func (r *CategoryRepository) Delete(id int) error {
	result, err := r.db.Exec(
		"UPDATE category SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL",
		id,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// Update updates an existing category in the database
func (r *CategoryRepository) Update(category models.Category) (models.Category, error) {
	var deletedAt sql.NullTime
	err := r.db.QueryRow(
		"UPDATE category SET name = $1, description = $2 WHERE id = $3 AND deleted_at IS NULL RETURNING id, name, description, deleted_at",
		category.Name, category.Description, category.ID,
	).Scan(&category.ID, &category.Name, &category.Description, &deletedAt)

	if err != nil {
		return models.Category{}, err
	}

	if deletedAt.Valid {
		category.DeletedAt = timestamppb.New(deletedAt.Time)
	}

	return category, nil
}
