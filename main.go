package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"kasir-api/docs"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	httpSwagger "github.com/swaggo/http-swagger"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Category represents a category in the cashier system
type Category struct {
	ID          int                    `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	DeletedAt   *timestamppb.Timestamp `json:"deleted_at"`
}

// Response represents the standardized API response format
type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// WriteJSON is a helper to write JSON responses
func WriteJSON(w http.ResponseWriter, status int, res Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(res)
}

// deleteCategory godoc
// @Summary      Delete a category
// @Description  Soft delete a category by ID
// @Tags         category
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Category ID"
// @Success      200  {object}  Response
// @Failure      400  {object}  Response
// @Failure      404  {object}  Response
// @Failure      500  {object}  Response
// @Router       /category/{id} [delete]
func deleteCategory(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// get id
	idStr := strings.TrimPrefix(r.URL.Path, "/api/category/")

	// ganti id int
	id, err := strconv.Atoi(idStr)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, Response{
			Status:  "failed",
			Message: "Invalid Category ID",
		})
		return
	}

	// Soft delete: set deleted_at timestamp
	result, err := db.Exec(
		"UPDATE category SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL",
		id,
	)

	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, Response{
			Status:  "failed",
			Message: "Failed to delete category: " + err.Error(),
		})
		return
	}

	// Check if any row was affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, Response{
			Status:  "failed",
			Message: "Failed to verify deletion: " + err.Error(),
		})
		return
	}

	if rowsAffected == 0 {
		WriteJSON(w, http.StatusNotFound, Response{
			Status:  "failed",
			Message: "Category not found",
		})
		return
	}

	WriteJSON(w, http.StatusOK, Response{
		Status:  "success",
		Message: "Category deleted successfully",
	})
}

// updateCategory godoc
// @Summary      Update a category
// @Description  Update a category by ID
// @Tags         category
// @Accept       json
// @Produce      json
// @Param        id        path      int       true  "Category ID"
// @Param        category  body      Category  true  "Category Data"
// @Success      200       {object}  Response{data=Category}
// @Failure      400       {object}  Response
// @Failure      404       {object}  Response
// @Failure      500       {object}  Response
// @Router       /category/{id} [put]
func updateCategory(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// get id dari request
	idStr := strings.TrimPrefix(r.URL.Path, "/api/category/")

	// ganti int
	id, err := strconv.Atoi(idStr)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, Response{
			Status:  "failed",
			Message: "Invalid Category ID",
		})
		return
	}

	// get data dari request
	var updateCategory Category
	err = json.NewDecoder(r.Body).Decode(&updateCategory)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, Response{
			Status:  "failed",
			Message: "Invalid request body",
		})
		return
	}

	// Fetch existing category first
	var existingCategory Category
	var existingDeletedAt sql.NullTime
	err = db.QueryRow(
		"SELECT id, name, description, deleted_at FROM category WHERE id = $1 AND deleted_at IS NULL",
		id,
	).Scan(&existingCategory.ID, &existingCategory.Name, &existingCategory.Description, &existingDeletedAt)

	if err == sql.ErrNoRows {
		WriteJSON(w, http.StatusNotFound, Response{
			Status:  "failed",
			Message: "Category not found",
		})
		return
	}

	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, Response{
			Status:  "failed",
			Message: "Failed to fetch category: " + err.Error(),
		})
		return
	}

	// Merge: only update fields that are provided in request body
	if updateCategory.Name != "" {
		existingCategory.Name = updateCategory.Name
	}
	if updateCategory.Description != "" {
		existingCategory.Description = updateCategory.Description
	}

	// Update category di database
	var deletedAt sql.NullTime
	err = db.QueryRow(
		"UPDATE category SET name = $1, description = $2 WHERE id = $3 AND deleted_at IS NULL RETURNING id, name, description, deleted_at",
		existingCategory.Name, existingCategory.Description, id,
	).Scan(&existingCategory.ID, &existingCategory.Name, &existingCategory.Description, &deletedAt)

	if err == sql.ErrNoRows {
		WriteJSON(w, http.StatusNotFound, Response{
			Status:  "failed",
			Message: "Category not found",
		})
		return
	}

	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, Response{
			Status:  "failed",
			Message: "Failed to update category: " + err.Error(),
		})
		return
	}

	if deletedAt.Valid {
		existingCategory.DeletedAt = timestamppb.New(deletedAt.Time)
	}

	WriteJSON(w, http.StatusOK, Response{
		Status:  "success",
		Message: "Category updated successfully",
		Data:    existingCategory,
	})
}

// getCategoryByID godoc
// @Summary      Get a category by ID
// @Description  Get a category by its ID
// @Tags         category
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Category ID"
// @Success      200  {object}  Response{data=Category}
// @Failure      400  {object}  Response
// @Failure      404  {object}  Response
// @Failure      500  {object}  Response
// @Router       /category/{id} [get]
func getCategoryByID(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Parse ID dari URL path
	// URL: /api/category/123 -> ID = 123
	idStr := strings.TrimPrefix(r.URL.Path, "/api/category/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, Response{
			Status:  "failed",
			Message: "Invalid Category ID",
		})
		return
	}

	// Query category dari database
	var c Category
	var deletedAt sql.NullTime
	err = db.QueryRow(
		"SELECT id, name, description, deleted_at FROM category WHERE id = $1 AND deleted_at IS NULL",
		id,
	).Scan(&c.ID, &c.Name, &c.Description, &deletedAt)

	if err == sql.ErrNoRows {
		// Kalau tidak found
		WriteJSON(w, http.StatusNotFound, Response{
			Status:  "failed",
			Message: "Category not found",
		})
		return
	}

	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, Response{
			Status:  "failed",
			Message: "Failed to fetch category: " + err.Error(),
		})
		return
	}

	if deletedAt.Valid {
		c.DeletedAt = timestamppb.New(deletedAt.Time)
	}

	WriteJSON(w, http.StatusOK, Response{
		Status:  "success",
		Message: "Category retrieved successfully",
		Data:    c,
	})
}

// @title           Kasir API
// @version         1.0
// @description     This is a sample server for a Cashier System.
// @BasePath        /api

func main() {
	// load .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, using system environment variables")
	}

	portStr := os.Getenv("PORT")
	if portStr == "" {
		portStr = "8080"
	}

	// Dynamic Swagger Host
	docs.SwaggerInfo.Host = "localhost:" + portStr

	// connect to DB
	connStr := os.Getenv("DATABASE_URL")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error opening database connection:", err)
	}
	defer db.Close()

	// check connection
	err = db.Ping()
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}

	fmt.Println("Successfully connected to database!")

	// {{host}}/health
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, Response{
			Status:  "success",
			Message: "API Running",
		})
	})

	// Swagger
	http.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	// Routes
	http.HandleFunc("/api/category/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			getCategoryByID(w, r, db)
		case "PUT":
			updateCategory(w, r, db)
		case "DELETE":
			deleteCategory(w, r, db)
		default:
			WriteJSON(w, http.StatusMethodNotAllowed, Response{
				Status:  "failed",
				Message: "Method not allowed",
			})
		}
	})

	http.HandleFunc("/api/category", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			getCategories(w, r, db)
		case "POST":
			createCategory(w, r, db)
		default:
			WriteJSON(w, http.StatusMethodNotAllowed, Response{
				Status:  "failed",
				Message: "Method not allowed",
			})
		}
	})

	fmt.Println("Server running on http://localhost:" + portStr)
	err = http.ListenAndServe(":"+portStr, nil)
	if err != nil {
		fmt.Println("Error running server:", err)
	}
}

// getCategories godoc
// @Summary      Get all categories
// @Description  Get a list of all active categories
// @Tags         category
// @Accept       json
// @Produce      json
// @Success      200  {object}  Response{data=[]Category}
// @Failure      500  {object}  Response
// @Router       /category [get]
func getCategories(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	rows, err := db.Query("SELECT id, name, description, deleted_at FROM category WHERE deleted_at IS NULL")
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, Response{
			Status:  "failed",
			Message: "Failed to fetch categories: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var c Category
		var deletedAt sql.NullTime
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &deletedAt); err != nil {
			WriteJSON(w, http.StatusInternalServerError, Response{
				Status:  "failed",
				Message: "Failed to scan category: " + err.Error(),
			})
			return
		}
		if deletedAt.Valid {
			c.DeletedAt = timestamppb.New(deletedAt.Time)
		}
		categories = append(categories, c)
	}

	WriteJSON(w, http.StatusOK, Response{
		Status:  "success",
		Message: "Categories retrieved successfully",
		Data:    categories,
	})
}

// createCategory godoc
// @Summary      Create a new category
// @Description  Create a new category with name and description
// @Tags         category
// @Accept       json
// @Produce      json
// @Param        category  body      Category  true  "Category Data"
// @Success      201       {object}  Response{data=Category}
// @Failure      400       {object}  Response
// @Failure      500       {object}  Response
// @Router       /category [post]
func createCategory(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// baca data dari request
	var categoryBaru Category
	err := json.NewDecoder(r.Body).Decode(&categoryBaru)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, Response{
			Status:  "failed",
			Message: "Invalid request body",
		})
		return
	}

	// simpan ke database
	var deletedAt sql.NullTime
	err = db.QueryRow(
		"INSERT INTO category (name, description) VALUES ($1, $2) RETURNING id, deleted_at",
		categoryBaru.Name, categoryBaru.Description,
	).Scan(&categoryBaru.ID, &deletedAt)

	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, Response{
			Status:  "failed",
			Message: "Failed to save category: " + err.Error(),
		})
		return
	}

	if deletedAt.Valid {
		categoryBaru.DeletedAt = timestamppb.New(deletedAt.Time)
	}

	WriteJSON(w, http.StatusCreated, Response{
		Status:  "success",
		Message: "Category created successfully",
		Data:    categoryBaru,
	})
}
